# AI-LOG

Catat setiap penggunaan AI yang menghasilkan kode. Jika tidak memakai AI sama sekali, tulis satu baris: **tidak memakai AI**.

Log yang jujur tidak mengurangi nilai, sebanyak apa pun AI dipakai. Yang mengurangi nilai adalah log yang tidak sesuai dengan kode, atau isi log yang tidak bisa dijelaskan.

## (Nama fitur atau fungsi, misalnya: Deteksi bentrok slot, B5)
- Prompt: (ringkas, apa yang kamu minta)
- Dipakai: (bagian kode yang kamu ambil, misalnya nama fungsi dan file)
- Diubah: (apa yang kamu ubah atau perbaiki dari hasil AI, dan kenapa)

## (Entri berikutnya)
- Prompt:
- Dipakai:
- Diubah:

## (Entri 1)
- Prompt: Bagaimana cara membuat penyimpanan data sumber daya dan peminjaman tanpa menggunakan database di Go?
- Dipakai:Struktur Store pada file main.go
- Diubah: saya memperbaiki kode yang awalnya saya samakan dengan struct yang digunakan dan menambahkan sync.Mutex untuk mengamankan akses dataseperti biasanya sehingga saya ubah menjadi :
type Store struct {
	mu               sync.Mutex
	sumberDaya       []SumberDaya
	peminjaman       []Peminjaman
	nextSumberDayaID int
	nextPeminjamanID int
}
