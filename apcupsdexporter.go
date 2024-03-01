// Package apcupsdexporter provides the Exporter type used in the
// apcupsd_exporter Prometheus exporter.
package apcupsdexporter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/mdlayher/apcupsd"
	"github.com/prometheus/client_golang/prometheus"
)

// TODO(mdlayher): rework this with metricslite.

const (
	// namespace is the top-level namespace for this apcupsd exporter.
	namespace = "apcupsd"
)

// An Exporter is a Prometheus exporter for apcupsd metrics.
// It wraps all apcupsd metrics collectors and provides a single global
// exporter which can serve metrics.
//
// It implements the prometheus.Collector interface in order to register
// with Prometheus.
type Exporter struct {
	clientFn ClientFunc
	logger   log.Logger

	Up *prometheus.Desc
}

var _ prometheus.Collector = &Exporter{}

// A ClientFunc is a function which can return an apcupsd NIS client.
// ClientFuncs are invoked on each Prometheus scrape, so that connections
// can be short-lived and less likely to time out or fail.
type ClientFunc func(ctx context.Context) (*apcupsd.Client, error)

// New creates a new Exporter which collects metrics by creating a apcupsd
// client using the input ClientFunc.
func New(fn ClientFunc, logger log.Logger) *Exporter {
	return &Exporter{
		clientFn: fn,
		logger:   logger,
		Up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"The status of the collector.",
			nil,
			nil,
		),
	}
}

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// Don't care, we'll get this error on Collect.
	_ = e.withCollectors(func(cs []prometheus.Collector) {
		for _, c := range cs {
			c.Describe(ch)
		}
	})
	ch <- e.Up
}

// Collect sends the collected metrics from each of the collectors to
// prometheus.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	err := e.withCollectors(func(cs []prometheus.Collector) {
		for _, c := range cs {
			c.Collect(ch)
		}
	})
	// This is a hack but it allows us to report failure to dial without
	// reworking significant portions of the code.
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.Up, prometheus.GaugeValue, 0)
		level.Error(e.logger).Log("err", err)
		// ch <- prometheus.NewInvalidMetric(NewUPSCollector(nil).Info, err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.Up, prometheus.GaugeValue, 1)
}

// withCollectors sets up an apcupsd client and creates a set of prometheus
// collectors.  It invokes the input closure and then cleans up after the
// closure returns.
func (e *Exporter) withCollectors(fn func(cs []prometheus.Collector)) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := e.clientFn(ctx)
	if err != nil {
		return fmt.Errorf("error creating apcupsd client: %v", err)
	}
	defer c.Close()

	cs := []prometheus.Collector{
		NewUPSCollector(c),
	}

	fn(cs)

	return nil
}
