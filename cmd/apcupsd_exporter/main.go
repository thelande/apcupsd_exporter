// Command apcupsd_exporter provides a Prometheus exporter for apcupsd.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mdlayher/apcupsd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	apcupsdexporter "github.com/thelande/apcupsd_exporter"
)

const (
	exporterName  = "apcupsd_exporter"
	exporterTitle = "apcupsd Exporter"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	webConfig   = webflag.AddFlags(kingpin.CommandLine, ":9162")
	apcupsdAddr = kingpin.Flag(
		"apcupsd.addr",
		"address of apcupsd Network Information Server (NIS)",
	).Default(":3551").String()
	apcupsdNetwork = kingpin.Flag(
		"apcupsd.network",
		`network of apcupsd Network Information Server (NIS): typically "tcp", "tcp4", or "tcp6"`,
	).Default("tcp").String()

	logger log.Logger
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print(exporterName))
	kingpin.Parse()

	logger = promlog.New(promlogConfig)
	level.Info(logger).Log("msg", fmt.Sprintf("Starting %s", exporterName), "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	if *apcupsdAddr == "" {
		level.Error(logger).Log("msg", "address of apcupsd Network Information Server (NIS) must be specified with '--apcupsd.addr' flag")
		os.Exit(1)
	}

	fn := newClient(*apcupsdNetwork, *apcupsdAddr)

	registry := prometheus.NewRegistry()
	registry.MustRegister(apcupsdexporter.New(fn, logger))

	landingConfig := web.LandingConfig{
		Name:        exporterTitle,
		Description: "Prometheus apcupsd Exporter",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{
				Address: *metricsPath,
				Text:    "Metrics",
			},
		},
	}
	landingPage, err := web.NewLandingPage(landingConfig)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.Handle("/", landingPage)

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "HTTP listener stopped", "error", err)
		os.Exit(1)
	}
}

func newClient(network, addr string) apcupsdexporter.ClientFunc {
	return func(ctx context.Context) (*apcupsd.Client, error) {
		return apcupsd.DialContext(ctx, network, addr)
	}
}
