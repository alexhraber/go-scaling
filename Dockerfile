# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.8
ARG MODULE=module-004-slices-maps-and-structs
ARG ENTRYPOINT_FILE=main.go

FROM gcr.io/distroless/static-debian12:nonroot AS runtime-base

WORKDIR /app
USER nonroot:nonroot

FROM golang:${GO_VERSION}-bookworm AS build

ARG MODULE
ARG ENTRYPOINT_FILE
ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

COPY go.mod ./
COPY ${MODULE}/${ENTRYPOINT_FILE} ./main.go

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/app ./main.go

FROM runtime-base AS app

COPY --from=build /out/app /app/app
ENTRYPOINT ["/app/app"]
