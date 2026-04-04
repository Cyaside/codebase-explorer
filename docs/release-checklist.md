# Release Checklist

## Sebelum Build

- `go test ./... -timeout 60s`
- `go run ./cmd/codearch doctor`
- `go run ./cmd/codearch analyze ./testdata/sample-repo --deterministic-only`
- `go run ./cmd/codearch open --no-browser`
- `go run ./cmd/codearch export --output ./out/release-smoke.zip`

## Build Targets

- Windows amd64
- macOS arm64
- Linux amd64

## Verifikasi Manual

- bundle terbaru punya `README.md`, `changes/`, `data/`, dan `ui/`
- viewer lokal tetap terbuka dari file system
- `cache clear` tidak menghapus `out/`
- `export` menghasilkan `.zip` yang bisa dibuka

## Catatan Release

- jaga binary tetap single-file
- jangan tambahkan dependency berat hanya untuk packaging
- kalau ada fitur eksperimen, jangan biarkan itu memblokir rilis utama
