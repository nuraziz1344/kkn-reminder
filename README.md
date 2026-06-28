# Bot Reminder KKN (WhatsApp)

Bot WhatsApp sederhana yang mengirim pengingat *checkpoint* KKN ke grup secara otomatis
setiap hari pada **06:00, 12:00, 18:00, 22:00, dan 23:45 WIB**. Setiap pengingat menandai
(@mention) seluruh anggota grup.

Dibangun dengan [whatsmeow](https://github.com/tulir/whatsmeow).

## Prasyarat

- Go 1.24+
- Compiler C (gcc) — diperlukan oleh driver SQLite (`CGO_ENABLED=1`)

## Build

```bash
go mod tidy
CGO_ENABLED=1 go build -o KKN
```

## Cara pakai

### 1. Login & cari JID grup

Jalankan dengan flag `-list`. Pertama kali akan muncul **QR code** di terminal — scan lewat
WhatsApp di HP: **Setelan > Perangkat Tertaut > Tautkan Perangkat**.

```bash
./KKN -list
```

Setelah login, bot menampilkan semua grup beserta JID-nya. Salin JID grup KKN (diakhiri
`@g.us`) ke `config.json`.

### 2. Atur konfigurasi (`config.json`)

```json
{
  "group_jid": "1234567890-1234567890@g.us",
  "timezone": "Asia/Jakarta",
  "message": "🔔 Reminder KKN! Waktunya checkpoint..."
}
```

- `group_jid` — JID grup tujuan (dari langkah 1)
- `timezone` — zona waktu (`Asia/Jakarta` = WIB, `Asia/Makassar` = WITA, `Asia/Jayapura` = WIT)
- `message` — isi pesan pengingat (mention anggota ditambahkan otomatis)

### 3. Uji coba (opsional)

Kirim satu pengingat sekarang juga untuk memastikan semuanya jalan:

```bash
./KKN -test
```

### 4. Jalankan

```bash
./KKN
```

Bot akan terus berjalan dan mengirim pengingat sesuai jadwal. Sesi login tersimpan di
`session.db`, jadi tidak perlu scan QR lagi pada run berikutnya.

## Tetap berjalan di latar belakang

Agar pengingat selalu terkirim, bot harus tetap hidup. Contoh dengan `tmux`:

```bash
tmux new -s kkn-bot
./KKN
# tekan Ctrl+B lalu D untuk detach
```

Atau buat layanan systemd (user service) untuk auto-start.

## Jadwal pengingat

Waktu dapat diubah di `scheduler.go` (variabel `checkpointSchedules`, format cron `menit jam * * *`).
Default:

| Waktu  | Cron          |
|--------|---------------|
| 06:00  | `0 6 * * *`   |
| 12:00  | `0 12 * * *`  |
| 18:00  | `0 18 * * *`  |
| 22:00  | `0 22 * * *`  |
| 23:45  | `45 23 * * *` |
