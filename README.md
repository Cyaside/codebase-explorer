# Codebase Explorer

Codebase Explorer adalah CLI local-first untuk membantu memahami codebase asing dengan cepat lewat structured analysis bundle.

Saat ini tool ini fokus pada deterministic repository orientation:
- scan repo lokal dengan ignore handling
- deteksi bahasa utama, entry point, dan module penting
- ranking hotspot dan dependency risk ringan
- reading path awal untuk onboarding
- optional support-file correlation untuk issue export atau changelog lokal
- output bundle reusable dalam format Markdown, JSON, Mermaid, dan viewer lokal statis

## Menjalankan

```bash
go run ./cmd/codearch doctor
go run ./cmd/codearch analyze <repo-path> --deterministic-only
go run ./cmd/codearch analyze <repo-path> --support ./issues.json --changelog ./CHANGELOG.md
go run ./cmd/codearch open --no-browser
```

## Output Bundle

Hasil analisis ditulis ke folder `out/` dan berisi:
- `README.md` sebagai pintu masuk hasil
- `overview/`, `architecture/`, `hotspots/`, `dependencies/`, dan `reading-path/`
- `changes/` untuk korelasi changelog atau issue export saat support files diberikan
- `architecture/module-graph.mmd` dan `dependencies/dependency-graph.mmd`
- `ui/index.html` untuk viewer lokal
- `data/analysis.json`, `data/metrics.json`, `data/files.json`, `data/modules.json`, dan `data/contract.json`
- `changes/issue-correlation.json` untuk hasil machine-readable change awareness

`out/` juga sekarang dipruning otomatis agar hanya menyimpan bundle terbaru dalam jumlah terbatas.

Support files tetap opsional. Kalau file issue/changelog tidak diberikan atau tidak valid, hasil utama deterministic tetap jadi dan bundle tetap ditulis.

## Fokus Saat Ini

Fase fondasi analyzer, synthesis AI opsional, dan visual report inti sudah aktif. Fokus berikutnya adalah memperkaya kualitas konsumsi hasil tanpa mengorbankan local-first workflow dan efisiensi runtime.
