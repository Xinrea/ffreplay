# FFXIV Replay

## Floor image size

1000 pixel = 40m

## Build

### Temporary WorldMarker API

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o markerserver ./cmd/apiserver/apiserver.go
```
