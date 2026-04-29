[tonyhsie/qBittorrentBlockXunlei](https://github.com/tonyhsie/qBittorrentBlockXunlei) rewritten in Go with WebUI.

## Usage

```
Usage of qbx:
  -addr string
        qBittorrent WebUI address (default "http://127.0.0.1:8080")
  -conf string
        config directory
  -password string
        qBittorrent WebUI password
  -username string
        qBittorrent WebUI username
  -webui string
        WebUI address (default ":2386")
```

## Build

```sh
pnpm install
pnpm bundle
go build -ldflags='-s -w' -trimpath
```

## Credits

- [fluentui-emoji](https://github.com/microsoft/fluentui-emoji) for favicon.
