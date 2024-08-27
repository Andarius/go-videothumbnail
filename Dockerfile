FROM golang:alpine AS builder

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

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.version=0.1.0" -o /go/bin/server


FROM alpine:3.17

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/server /go/bin/server

RUN apk add --no-cache ffmpeg
USER appuser:appuser

LABEL version="0.1.0"
LABEL organization="Obitrain"


EXPOSE 8080

ENTRYPOINT ["/go/bin/server"]
