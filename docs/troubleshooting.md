# Troubleshooting

## `doctor` Menunjukkan Provider Fail

Penyebab umum:
- `CODEARCH_PROVIDER` sudah diisi, tetapi `CODEARCH_MODEL` belum
- `CODEARCH_API_KEY` belum diisi
- `CODEARCH_BASE_URL` belum diisi untuk `openai-compatible`

Solusi:
- cek environment variable yang aktif
- jalankan `go run ./cmd/codearch doctor` lagi setelah diperbaiki

Kalau kamu memakai workbench:
- cek profile yang sedang dipilih
- pastikan `model`, `api_key`, dan `base_url` profile terisi sesuai provider
- simpan profile lagi lalu ulangi analisis

## AI Status `fallback`

Artinya deterministic analysis tetap sukses, tetapi synthesis AI gagal atau tidak valid.

Langkah cek:
- pastikan API key dan model benar
- cek apakah provider mengembalikan response yang valid
- untuk OpenRouter atau Mistral lewat mode compatible, pastikan base URL profile benar
- ulangi dengan `--deterministic-only` bila ingin memastikan jalur baseline tetap sehat

## Workbench Tidak Bisa Menganalisis GitHub URL

Itu memang perilaku saat ini.

Workbench sengaja tetap local-first dan lightweight, jadi URL remote belum di-clone otomatis. Solusinya:
- clone repo lebih dulu ke lokal
- jalankan analisis dengan local path hasil clone

## Changes Report Kosong

Penyebab umum:
- support file tidak diberikan
- support file tidak valid
- support file valid tetapi tidak cukup cocok dengan nama file/folder/module repo

Solusi:
- berikan `--support`, `--issues`, atau `--changelog`
- gunakan file yang menyebut path, module, atau fitur dengan lebih jelas

## Bundle Terasa Berat

Penyebab umum:
- repo sangat besar
- support files terlalu banyak
- viewer membawa data bundle yang juga besar

Solusi:
- jalankan deterministic-only dulu
- tambah ignore pattern yang relevan
- bagi analisis menjadi repo atau subdirectory yang lebih fokus bila perlu

## Cache Tidak Ingin Dipakai

Nonaktifkan cache untuk satu sesi:

```bash
export CODEARCH_CACHE=false
```

Atau bersihkan total:

```bash
go run ./cmd/codearch cache clear
```
