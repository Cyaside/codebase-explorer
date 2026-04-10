# Codebase Explorer

Codebase Explorer adalah CLI dan local web workbench untuk membantu memahami codebase asing dengan cepat lewat structured analysis bundle.

Tool ini memakai deterministic analysis sebagai baseline dan bisa dinaikkan ke full-AI repository exploration saat provider AI dikonfigurasi:
- scan repo lokal dengan ignore handling
- deteksi bahasa utama, entry point, dan module penting
- ranking hotspot dan dependency risk ringan
- reading path awal untuk onboarding
- optional support-file correlation untuk issue export atau changelog lokal
- optional full-AI mode berbasis instruction pack `.agents/`
- output bundle reusable dalam format Markdown, JSON, Mermaid, dan viewer/workbench lokal

## Quickstart

```bash
go run ./cmd/codearch doctor
go run ./cmd/codearch analyze <repo-path> --deterministic-only
go run ./cmd/codearch analyze <repo-path> --full-ai
go run ./cmd/codearch start --no-browser
go run ./cmd/codearch analyze <repo-path> --support ./issues.json --changelog ./CHANGELOG.md
go run ./cmd/codearch open --no-browser
go run ./cmd/codearch export --output ./out/latest-bundle.zip
go run ./cmd/codearch cache clear
```

Panduan langkah cepat yang lebih lengkap ada di [docs/quickstart.md](docs/quickstart.md).

## Commands

- `codearch doctor`
  Validasi config, output root, cache root, dan provider setup.
- `codearch analyze <repo-path>`
  Menjalankan scan, deterministic analysis, support-file correlation opsional, AI synthesis opsional, dan full-AI exploration bila `--full-ai` aktif.
- `codearch start [--addr <host:port>] [--no-browser]`
  Menjalankan local web workbench untuk membuka satu project aktif, menyimpan workspace lokal, memilih koneksi provider, menjalankan full-AI analysis, dan membaca bundle dengan UI dark control-plane. Ini adalah command yang direkomendasikan untuk pemakaian harian.
- `codearch serve [--addr <host:port>] [--no-browser]`
  Alias lama untuk `codearch start`.
- `codearch open [bundle-path] [--no-browser]`
  Membuka viewer bundle terbaru atau bundle yang dipilih.
- `codearch export [bundle-path] [--output <zip-path>]`
  Mengekspor bundle yang sudah ada ke file `.zip` tanpa analisis ulang.
- `codearch cache clear`
  Menghapus cache filesystem dengan aman tanpa menyentuh bundle di `out/`.

## Environment

Variabel environment yang didukung:

- `CODEARCH_OUTPUT_ROOT`
  Override root output bundle. Default: `out/`
- `CODEARCH_OUTPUT_KEEP`
  Jumlah bundle terbaru yang dipertahankan. Default: `10`
- `CODEARCH_CACHE`
  Aktif/nonaktifkan cache filesystem. Nilai: `true/false`
- `CODEARCH_CACHE_ROOT`
  Override lokasi cache filesystem
- `CODEARCH_PROVIDER`
  Provider AI opsional, misalnya `openai` atau `openai-compatible`
- `CODEARCH_MODEL`
  Model provider AI
- `CODEARCH_API_KEY`
  API key provider AI
- `CODEARCH_BASE_URL`
  Base URL untuk provider `openai-compatible`

Environment tetap cocok untuk satu default provider. Kalau butuh beberapa API key sekaligus, gunakan `codearch start` lalu simpan connection profile lokal di browser untuk OpenAI, OpenRouter, Mistral, atau endpoint compatible lain. Profile itu hanya dipakai per run dan tidak ikut ditulis ke bundle output.

Contoh Mistral lewat provider openai-compatible:

```powershell
$env:CODEARCH_PROVIDER="openai-compatible"
$env:CODEARCH_BASE_URL="https://api.mistral.ai/v1"
$env:CODEARCH_MODEL="mistral-small-latest"
$env:CODEARCH_API_KEY="..."
go run ./cmd/codearch analyze <repo-path> --full-ai
```

## Output Bundle

Hasil analisis ditulis ke folder `out/` dan berisi:
- `README.md` sebagai pintu masuk hasil
- `overview/`, `architecture/`, `hotspots/`, `dependencies/`, dan `reading-path/`
- `changes/` untuk korelasi changelog atau issue export saat support files diberikan
- `architecture/module-graph.mmd` dan `dependencies/dependency-graph.mmd`
- `ui/index.html` untuk viewer lokal
- `data/analysis.json`, `data/metrics.json`, `data/files.json`, `data/modules.json`, dan `data/contract.json`
- `data/full-ai-plan.json`, `data/full-ai-evidence.json`, `data/full-ai-functions.json`, `data/full-ai-execution.json`, dan `data/full-ai-meta.json`
- `changes/issue-correlation.json` untuk hasil machine-readable change awareness

`out/` juga sekarang dipruning otomatis agar hanya menyimpan bundle terbaru dalam jumlah terbatas.

Support files tetap opsional. Kalau file issue/changelog tidak diberikan atau tidak valid, hasil utama deterministic tetap jadi dan bundle tetap ditulis.

## Workbench

Workbench adalah local webapp ringan yang tetap jalan dari binary Go yang sama, tanpa Electron atau backend berat tambahan. Workbench ini cocok untuk:
- membuka satu project aktif lewat saved workspace lokal
- menyimpan beberapa connection profile secara lokal di browser
- menjalankan full-AI analysis dari koneksi provider terpilih
- langsung membaca dashboard, summary, architecture, flowchart, issue tracking, dan recommendations
- menampilkan flowchart project langsung di tab interaktif, bukan hanya source diagram mentah
- melihat status full-AI, jumlah evidence yang dibaca, function outputs, dan verified counts

Untuk sekarang jalur local-first tetap diprioritaskan, jadi input repository GitHub URL belum di-clone otomatis. Gunakan local checkout path saat menjalankan analisis dari workbench.

## Build

Single binary tetap jadi jalur distribusi utama. Contoh build:

```bash
go build -o ./dist/codearch ./cmd/codearch
GOOS=windows GOARCH=amd64 go build -o ./dist/codearch-windows-amd64.exe ./cmd/codearch
GOOS=darwin GOARCH=arm64 go build -o ./dist/codearch-darwin-arm64 ./cmd/codearch
GOOS=linux GOARCH=amd64 go build -o ./dist/codearch-linux-amd64 ./cmd/codearch
```

Release checklist ringkas ada di [docs/release-checklist.md](docs/release-checklist.md).

## Troubleshooting

Kalau run tidak sesuai harapan, lihat [docs/troubleshooting.md](docs/troubleshooting.md).

Masalah yang paling umum:
- provider belum dikonfigurasi, jadi AI otomatis `disabled` atau `fallback`
- support file tidak valid, jadi change-awareness ditulis sebagai partial result
- repo cukup besar, sehingga CLI memberi warning performa dan bundle bisa lebih berat

## Status

Deterministic analyzer, AI synthesis opsional, full-AI exploration, visual workbench, change-awareness, cache filesystem, `export`, dan `cache clear` sudah aktif. Produk sekarang sudah bisa dipakai end-to-end sebagai local-first repository orientation tool dengan optional deep AI pass.
