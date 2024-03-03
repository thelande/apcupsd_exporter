FROM --platform=${BUILDPLATFORM} golang:1.21-alpine AS builder
LABEL maintainer="Tom Helander <thomas.helander@gmail.com>"

RUN set -eux; \
	apk update; \
	apk upgrade --no-cache; \
	apk add --no-cache \
		curl \
		git \
		make \
	; \
	apk cache purge

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
RUN make GOOS=$TARGETOS GOARCH=$TARGETARCH build

FROM alpine:3.18.4
LABEL maintainer="Tom Helander <thomas.helander@gmail.com>"

WORKDIR /app

COPY --from=builder --chmod=0755 /src/output/apcupsd_exporter .

EXPOSE 9162

ENTRYPOINT ["/app/apcupsd_exporter"]
