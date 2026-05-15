# CWD Forum Backend

## Perintah Makefile

Proyek ini menggunakan `Makefile` untuk mempermudah operasional:

| Perintah | Deskripsi |
| :--- | :--- |
| `make run` | Menjalankan server aplikasi secara langsung. |
| `make migrate` | Menjalankan migrasi database (AutoMigrate) dan Seeder secara manual. |
| `make build` | Melakukan kompilasi aplikasi ke dalam direktori `bin/`. |
| `make lint` | Menjalankan pengecekan lint menggunakan `golangci-lint`. |
| `make fix` | Menjalankan auto-fix lint yang didukung oleh `golangci-lint`. |
| `make clean` | Menghapus file hasil build di direktori `bin/`. |

Pastikan binary `golangci-lint` sudah terpasang dan tersedia di `PATH` sebelum menjalankan `make lint` atau `make fix`.

## Alur Kerja Migrasi Database

Berbeda dengan setup standar, migrasi database di proyek ini dilakukan secara **manual** untuk mencegah perubahan skema yang tidak disengaja saat server dijalankan.

Jika Anda melakukan perubahan pada `internal/model/*.go` atau ingin mengisi ulang data awal (seeding), jalankan:
```bash
make migrate
```
