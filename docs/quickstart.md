# Quickstart

## Deterministic Only

1. Jalankan `go run ./cmd/codearch doctor`
2. Jalankan `go run ./cmd/codearch analyze <repo-path> --deterministic-only`
3. Buka hasil dengan `go run ./cmd/codearch open --no-browser`

## Dengan Support Files

1. Siapkan repo lokal
2. Siapkan file pendukung seperti `issues.json`, `issues.md`, atau `CHANGELOG.md`
3. Jalankan:

```bash
go run ./cmd/codearch analyze <repo-path> --issues ./issues.json --changelog ./CHANGELOG.md
```

4. Buka `changes/README.md` di bundle atau viewer lokal

## Dengan Provider AI

1. Set environment:

```bash
export CODEARCH_PROVIDER=openai
export CODEARCH_MODEL=gpt-4.1-mini
export CODEARCH_API_KEY=...
```

2. Jalankan `go run ./cmd/codearch doctor`
3. Jalankan `go run ./cmd/codearch analyze <repo-path>`

## Export Bundle

Arsipkan bundle terbaru tanpa analisis ulang:

```bash
go run ./cmd/codearch export --output ./out/latest-bundle.zip
```

## Clear Cache

Hapus cache filesystem tanpa menghapus bundle:

```bash
go run ./cmd/codearch cache clear
```
