# deeix-pd-mask: masks personal data in DEEIX Chat requests to LLM providers.
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/deeix-pd-mask .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/deeix-pd-mask /deeix-pd-mask
ENV PD_ADDR=:8090 GOMEMLIMIT=48MiB
USER 65534:65534
EXPOSE 8090
HEALTHCHECK --interval=60s --timeout=5s --start-period=5s CMD ["/deeix-pd-mask", "healthcheck"]
ENTRYPOINT ["/deeix-pd-mask"]
