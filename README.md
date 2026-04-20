# go-videothumbnail

Lightweight HTTP service that generates JPEG thumbnails from videos using FFmpeg.

## Requirements

- Go 1.22+
- FFmpeg and FFprobe installed

## Endpoints

### `GET /gen-thumb`

Generates a thumbnail from a video file.

| Parameter | Description |
|-----------|-------------|
| `path`    | Absolute path to the input video file |
| `output`  | Absolute path for the generated thumbnail |

Returns `201` with video dimensions:

```json
{"width": 1920, "height": 1080}
```

Returns `500` with error details on failure:

```json
{"error": "generate thumbnail for /path/to/video.mp4: ffmpeg failed: ..."}
```

### `GET /health`

Health check endpoint, returns `200`.

## Configuration

| Environment variable | Description | Required |
|---------------------|-------------|----------|
| `SENTRY_DSN`        | Sentry DSN for error tracking | No |
| `RELEASE_STAGE`     | Release identifier sent to Sentry | No |
| `SENTRY_TRACES_SAMPLE_RATE` | Sentry traces sample rate (default: `1.0`) | No |
| `SENTRY_PROFILES_SAMPLE_RATE` | Sentry profiles sample rate (default: `1.0`) | No |

## Sentry support

Sentry is optional and excluded from the default build. To enable it, use the `sentry` build tag:

```bash
# Local build with Sentry
go build -tags sentry -o server .

# Docker build with Sentry
docker build --build-arg BUILD_TAGS=sentry -t go-videothumbnail .
```

## Usage

```bash
# Run locally
go run main.go

# Docker
docker run --name thumb-gen \
    -v "$(pwd)/static/:/static" \
    -p 127.0.0.1:8080:8080 \
    -u "$(id -u):$(id -g)" \
    -d \
    andarius/go-videothumbnail:latest

# Generate a thumbnail
curl "http://127.0.0.1:8080/gen-thumb?path=/static/video.mp4&output=/static/thumb.jpeg"

# Stop
docker stop thumb-gen && docker rm thumb-gen
```