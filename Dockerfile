FROM --platform=${BUILDPLATFORM} golang:1.24-alpine AS builder
LABEL maintainer="Tom Helander <thomas.helander@gmail.com>"

RUN apk add make curl git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
RUN make GOOS=$TARGETOS GOARCH=$TARGETARCH build

FROM alpine:3.18.4
LABEL maintainer="Tom Helander <thomas.helander@gmail.com>"

WORKDIR /app
RUN set -eux; \
    apk update; \
    apk upgrade -v; \
    apk cache purge

COPY --from=builder --chmod=0755 /src/output/apcupsd_exporter .

EXPOSE 9162

ENTRYPOINT ["/app/apcupsd_exporter"]
