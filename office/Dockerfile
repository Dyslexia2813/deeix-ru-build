# deeix-office: MCP server for DEEIX Chat that makes Word (.docx) and Excel (.xlsx) files for the model.
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/deeix-office . && mkdir -p /out/data

FROM scratch
COPY --from=build /out/deeix-office /deeix-office
# The data volume takes this owner the first time it is mounted.
COPY --from=build --chown=65534:65534 /out/data /data
ENV OFFICE_ADDR=:8095 OFFICE_DIR=/data GOMEMLIMIT=40MiB
USER 65534:65534
EXPOSE 8095
VOLUME /data
HEALTHCHECK --interval=60s --timeout=5s --start-period=5s CMD ["/deeix-office", "healthcheck"]
ENTRYPOINT ["/deeix-office"]
