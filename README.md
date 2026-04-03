# Codebase Explorer

Codebase Explorer adalah CLI local-first untuk membantu memahami codebase asing dengan cepat lewat structured analysis bundle.

Saat ini tool ini fokus pada deterministic repository orientation:
- scan repo lokal dengan ignore handling
- deteksi bahasa utama, entry point, dan module penting
- ranking hotspot dan dependency risk ringan
- reading path awal untuk onboarding
- output bundle reusable dalam format Markdown dan JSON

## Menjalankan

```bash
go run ./cmd/codearch doctor
go run ./cmd/codearch analyze <repo-path> --deterministic-only
```

## Output Bundle

Hasil analisis ditulis ke folder `out/` dan berisi:
- `README.md` sebagai pintu masuk hasil
- `overview/`, `architecture/`, `hotspots/`, `dependencies/`, dan `reading-path/`
- `data/analysis.json`, `data/metrics.json`, `data/files.json`, `data/modules.json`, dan `data/contract.json`

## Fokus Saat Ini

Phase 1 membangun fondasi analyzer yang tetap berguna tanpa AI, sehingga hasil scan, heuristik deterministic, dan bundle output sudah bisa dipakai sendiri untuk orientasi awal repository.
