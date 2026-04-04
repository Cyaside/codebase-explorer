# Quickstart

## Deterministic Only

1. Jalankan `go run ./cmd/codearch doctor`
2. Jalankan `go run ./cmd/codearch analyze <repo-path> --deterministic-only`
3. Buka hasil dengan `go run ./cmd/codearch open --no-browser`

## Local Workbench

1. Jalankan `go run ./cmd/codearch start`
2. Masukkan local repo path di form workbench
3. Pilih connection profile:
   `Deterministic only`, `OpenAI`, `OpenRouter`, `Mistral`, atau custom compatible profile
4. Jalankan analisis dari UI dan pilih bundle hasilnya dari sidebar

Catatan:
- connection profile disimpan lokal di browser
- API key tidak ikut ditulis ke bundle
- GitHub URL belum di-clone otomatis; gunakan local checkout path

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

Alternatifnya, gunakan `go run ./cmd/codearch start` lalu simpan beberapa connection profile di workbench bila kamu ingin berganti API key atau vendor per analisis.

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
