# Go-vidthumb


Generate thumbnails from videos.

### How to use

```bash
# Start docker image
docker run --name thumb-gen \
    -v "$(pwd)/static/:/static" \
    -p 127.0.0.1:8080:8080 \
    -u "$(id -u):$(id -g)" \
    -d \
    andarius/go-vidthumbnail:latest

# Call endpoint
curl http://127.0.0.1:8080/gen-thumb\?path\=/static/video.mp4\&output\=/static/thumb.png

# Stop docker image
docker stop thumb-gen && docker rm thumb-gen
```
