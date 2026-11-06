# deeix-ocr: Tesseract (rus+eng) behind a small Go HTTP server for DEEIX Chat.
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/deeix-ocr .

FROM ubuntu:24.04
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       tesseract-ocr tesseract-ocr-rus tesseract-ocr-eng poppler-utils \
    && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man \
    && useradd --system --no-create-home --shell /usr/sbin/nologin ocr
COPY --from=build /out/deeix-ocr /usr/local/bin/deeix-ocr
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
ENV OCR_ADDR=:8004 OCR_LANG=rus+eng OCR_DPI=200 OCR_MAX_PAGES=20 OMP_THREAD_LIMIT=1 GOMEMLIMIT=32MiB
USER ocr
EXPOSE 8004
HEALTHCHECK --interval=60s --timeout=5s --start-period=10s CMD ["/usr/local/bin/deeix-ocr", "healthcheck"]
ENTRYPOINT ["/bin/sh", "/usr/local/bin/entrypoint.sh"]
