
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

type Peminjaman struct {
	ID           int    `json:"id"`
	SumberDayaID int    `json:"sumber_daya_id"`
	Peminjam     string `json:"peminjam"`
	Tanggal      string `json:"tanggal"`
	SlotMulai    int    `json:"slot_mulai"`
	SlotSelesai  int    `json:"slot_selesai"`
	Jumlah       int    `json:"jumlah"`
	Status       string `json:"status"`
}

type Store struct {
	mu               sync.Mutex
	sumberDaya       []SumberDaya
	peminjaman       []Peminjaman
	nextSumberDayaID int
	nextPeminjamanID int
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	store := &Store{}

	// B1: Endpoint sumber daya
	mux.HandleFunc("POST /sumber-daya", store.handleCreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", store.handleListSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", store.handleGetSumberDaya)

	// B2: Membuat peminjaman
	mux.HandleFunc("POST /peminjaman", store.handleCreatePeminjaman)

	// B3: Daftar dan detail peminjaman
	mux.HandleFunc("GET /peminjaman", store.handleListPeminjaman)
	mux.HandleFunc("GET /peminjaman/{id}", store.handleGetPeminjaman)

	return mux
}

// B4: Mengirim respons dalam format JSON.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// B4: Memeriksa format dan validitas tanggal YYYY-MM-DD.
func validTanggal(tanggal string) bool {
	t, err := time.Parse("2006-01-02", tanggal)
	return err == nil && t.Format("2006-01-02") == tanggal
}

// B1: Membuat sumber daya.
func (s *Store) handleCreateSumberDaya(w http.ResponseWriter, r *http.Request) {
	var data SumberDaya

	// B4: Memeriksa apakah body request berupa JSON valid.
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Format JSON tidak valid",
		})
		return
	}

	// B4: Menghapus spasi di awal dan akhir teks.
	data.Nama = strings.TrimSpace(data.Nama)
	data.Jenis = strings.TrimSpace(data.Jenis)

	// B4: Nama wajib diisi dan maksimal 50 karakter.
	if data.Nama == "" || len([]rune(data.Nama)) > 50 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Nama wajib diisi dan maksimal 50 karakter",
		})
		return
	}

	// B4: Jenis hanya boleh ruang atau alat.
	if data.Jenis != "ruang" && data.Jenis != "alat" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Jenis harus ruang atau alat",
		})
		return
	}

	// B4: Stok harus berupa bilangan bulat minimal 1.
	if data.Stok < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Stok minimal 1",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// B1: Membuat ID sumber daya secara otomatis.
	s.nextSumberDayaID++
	data.ID = s.nextSumberDayaID

	s.sumberDaya = append(s.sumberDaya, data)

	writeJSON(w, http.StatusCreated, data)
}

// B1: Menampilkan semua sumber daya.
func (s *Store) handleListSumberDaya(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// B1: Menghasilkan array kosong [] jika belum ada data.
	hasil := make([]SumberDaya, len(s.sumberDaya))
	copy(hasil, s.sumberDaya)

	writeJSON(w, http.StatusOK, hasil)
}

// B1: Menampilkan sumber daya berdasarkan ID.
func (s *Store) handleGetSumberDaya(w http.ResponseWriter, r *http.Request) {
	// B4: ID bukan angka positif menghasilkan 400.
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "ID harus berupa angka positif",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, data := range s.sumberDaya {
		if data.ID == id {
			writeJSON(w, http.StatusOK, data)
			return
		}
	}

	// B4: ID valid tetapi data tidak ditemukan menghasilkan 404.
	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "Sumber daya tidak ditemukan",
	})
}

// B2: Membuat peminjaman.
func (s *Store) handleCreatePeminjaman(w http.ResponseWriter, r *http.Request) {
	var p Peminjaman

	// B4: Memeriksa apakah body request berupa JSON valid.
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Format JSON tidak valid",
		})
		return
	}

	// B4: Menghapus spasi di awal dan akhir teks.
	p.Peminjam = strings.TrimSpace(p.Peminjam)
	p.Tanggal = strings.TrimSpace(p.Tanggal)

	// B4: Memvalidasi field wajib dan rentang slot 1 sampai 8.
	if p.SumberDayaID < 1 ||
		p.Peminjam == "" ||
		len([]rune(p.Peminjam)) > 50 ||
		p.SlotMulai < 1 || p.SlotMulai > 8 ||
		p.SlotSelesai < p.SlotMulai || p.SlotSelesai > 8 ||
		p.Jumlah < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Data peminjaman tidak valid",
		})
		return
	}

	// B4: Tanggal harus valid dan mengikuti format YYYY-MM-DD.
	if !validTanggal(p.Tanggal) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Format tanggal harus YYYY-MM-DD dan tanggal harus valid",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// B2: Mencari sumber daya yang akan dipinjam.
	var sumber *SumberDaya

	for i := range s.sumberDaya {
		if s.sumberDaya[i].ID == p.SumberDayaID {
			sumber = &s.sumberDaya[i]
			break
		}
	}

	// B4: sumber_daya_id yang tidak ditemukan menghasilkan 400,
	// sesuai kontrak ETS.
	if sumber == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Sumber daya tidak ditemukan",
		})
		return
	}

	// B4: Jumlah harus berada di antara 1 dan stok sumber daya.
	if p.Jumlah > sumber.Stok {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Jumlah melebihi stok sumber daya",
		})
		return
	}

	// B2: Membuat ID dan menetapkan status awal peminjaman.
	s.nextPeminjamanID++
	p.ID = s.nextPeminjamanID
	p.Status = "aktif"

	s.peminjaman = append(s.peminjaman, p)

	writeJSON(w, http.StatusCreated, p)
}

// B3: Menampilkan daftar peminjaman dengan filter opsional.
func (s *Store) handleListPeminjaman(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	sumberDayaID := strings.TrimSpace(query.Get("sumber_daya_id"))
	tanggal := strings.TrimSpace(query.Get("tanggal"))
	status := strings.TrimSpace(query.Get("status"))

	idFilter := 0

	// B4: Filter sumber_daya_id harus berupa angka positif.
	if sumberDayaID != "" {
		id, err := strconv.Atoi(sumberDayaID)
		if err != nil || id < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "sumber_daya_id harus berupa angka positif",
			})
			return
		}
		idFilter = id
	}

	// B4: Jika filter tanggal diisi, tanggal harus valid.
	if tanggal != "" && !validTanggal(tanggal) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Format tanggal harus YYYY-MM-DD dan tanggal harus valid",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// B3: Slice kosong memastikan hasil tanpa data berupa [].
	hasil := make([]Peminjaman, 0)

	for _, p := range s.peminjaman {
		if idFilter != 0 && p.SumberDayaID != idFilter {
			continue
		}

		if tanggal != "" && p.Tanggal != tanggal {
			continue
		}

		if status != "" && p.Status != status {
			continue
		}

		hasil = append(hasil, p)
	}

	// Urutan hasil mengikuti ID yang dibuat secara berurutan.
	writeJSON(w, http.StatusOK, hasil)
}

// B3: Menampilkan peminjaman berdasarkan ID.
func (s *Store) handleGetPeminjaman(w http.ResponseWriter, r *http.Request) {
	// B4: ID bukan angka positif menghasilkan 400.
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "ID harus berupa angka positif",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.peminjaman {
		if p.ID == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}

	// B4: ID valid tetapi peminjaman tidak ditemukan menghasilkan 404.
	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "Peminjaman tidak ditemukan",
	})
}

func main() {
	log.Println("Server berjalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", NewHandler()))
}
