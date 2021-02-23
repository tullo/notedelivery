FROM golang:1.16.0-alpine3.13 AS go-builder
ENV CGO_ENABLED 0
WORKDIR /build
COPY . .
WORKDIR /build/app/notedelivery
RUN go build -mod=vendor


FROM alpine:3.13.2
RUN apk --no-cache add ca-certificates
RUN addgroup -g 3000 -S app && adduser -u 100000 -S app -G app --no-create-home --disabled-password \
    && mkdir -p /app/badger.db && chown app:app /app/badger.db
USER 100000
WORKDIR /app
COPY --from=go-builder --chown=app:app /build/app/notedelivery/notedelivery /app/notedelivery
COPY --from=go-builder --chown=app:app /build/ui/static /app/ui/static
ENTRYPOINT ["/app/notedelivery"]
