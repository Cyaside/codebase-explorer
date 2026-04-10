# Quickstart

## Deterministic Only

1. Jalankan `go run ./cmd/codearch doctor`
2. Jalankan `go run ./cmd/codearch analyze <repo-path> --deterministic-only`
3. Buka hasil dengan `go run ./cmd/codearch open --no-browser`

## Local Workbench

1. Jalankan `go run ./cmd/codearch start`
2. Masukkan local repo path di form workbench
3. Pilih connection profile:
   `OpenAI`, `OpenRouter`, `Mistral`, atau custom compatible profile
4. Masukkan API key untuk profile yang dipakai
5. Jalankan analisis dari UI dan pilih bundle hasilnya dari sidebar

Catatan:
- connection profile disimpan lokal di browser
- API key tidak ikut ditulis ke bundle
- workbench menjalankan full-AI mode saat connection profile valid
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

1. Set environment untuk provider bawaan:

```bash
export CODEARCH_PROVIDER=openai
export CODEARCH_MODEL=gpt-4.1-mini
export CODEARCH_API_KEY=...
```

2. Jalankan `go run ./cmd/codearch doctor`
3. Jalankan `go run ./cmd/codearch analyze <repo-path>`

Alternatifnya, gunakan `go run ./cmd/codearch start` lalu simpan beberapa connection profile di workbench bila kamu ingin berganti API key atau vendor per analisis.

## Full-AI Mode

Full-AI mode membaca instruction pack `.agents/`, mengumpulkan evidence dari file high-signal, lalu menjalankan function outputs untuk summary, architecture, flowchart, issues, recommendations, dashboard, dan hotspot/dependency review.

Contoh CLI dengan Mistral:

```powershell
$env:CODEARCH_PROVIDER="openai-compatible"
$env:CODEARCH_BASE_URL="https://api.mistral.ai/v1"
$env:CODEARCH_MODEL="mistral-small-latest"
$env:CODEARCH_API_KEY="..."
go run ./cmd/codearch analyze <repo-path> --full-ai --ai-read-budget 24
```

Kalau ingin test tanpa memanggil provider:

```bash
go run ./cmd/codearch analyze <repo-path> --full-ai --deterministic-only
```

Output full-AI tersimpan di:
- `data/full-ai-plan.json`
- `data/full-ai-evidence.json`
- `data/full-ai-functions.json`
- `data/full-ai-execution.json`
- `data/full-ai-meta.json`

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
