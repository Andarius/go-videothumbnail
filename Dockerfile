FROM golang:1.22-alpine AS builder

ARG VERSION=dev

ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /go/delivery

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILD_TAGS=""
RUN go build -tags="${BUILD_TAGS}" -ldflags="-w -s -X main.version=${VERSION}" -o /go/bin/server


FROM alpine:3.21

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/server /go/bin/server

RUN apk add --no-cache ffmpeg
USER appuser:appuser

EXPOSE 8080

ENTRYPOINT ["/go/bin/server"]