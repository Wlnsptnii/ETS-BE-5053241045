
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

	// B1: Sumber daya
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

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// B1: Membuat sumber daya
func (s *Store) handleCreateSumberDaya(w http.ResponseWriter, r *http.Request) {
	var data SumberDaya

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Format JSON tidak valid",
		})
		return
	}

	data.Nama = strings.TrimSpace(data.Nama)
	data.Jenis = strings.TrimSpace(data.Jenis)

	if data.Nama == "" || len([]rune(data.Nama)) > 50 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Nama wajib diisi dan maksimal 50 karakter",
		})
		return
	}

	if data.Jenis != "ruang" && data.Jenis != "alat" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Jenis harus ruang atau alat",
		})
		return
	}

	if data.Stok < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Stok minimal 1",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextSumberDayaID++
	data.ID = s.nextSumberDayaID

	s.sumberDaya = append(s.sumberDaya, data)

	writeJSON(w, http.StatusCreated, data)
}

// B1: Menampilkan semua sumber daya
func (s *Store) handleListSumberDaya(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := make([]SumberDaya, len(s.sumberDaya))
	copy(data, s.sumberDaya)

	writeJSON(w, http.StatusOK, data)
}

// B1: Menampilkan sumber daya berdasarkan ID
func (s *Store) handleGetSumberDaya(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "Sumber daya tidak ditemukan",
	})
}

// B2: Membuat peminjaman
func (s *Store) handleCreatePeminjaman(w http.ResponseWriter, r *http.Request) {
	var p Peminjaman

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Format JSON tidak valid",
		})
		return
	}

	p.Peminjam = strings.TrimSpace(p.Peminjam)
	p.Tanggal = strings.TrimSpace(p.Tanggal)

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

	// Memastikan tanggal menggunakan format YYYY-MM-DD
	tanggal, err := time.Parse("2006-01-02", p.Tanggal)
	if err != nil || tanggal.Format("2006-01-02") != p.Tanggal {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Format tanggal harus YYYY-MM-DD dan tanggal harus valid",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Mencari sumber daya yang akan dipinjam
	var sumber *SumberDaya

	for i := range s.sumberDaya {
		if s.sumberDaya[i].ID == p.SumberDayaID {
			sumber = &s.sumberDaya[i]
			break
		}
	}

	if sumber == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Sumber daya tidak ditemukan",
		})
		return
	}

	// Memastikan jumlah tidak melebihi stok total
	if p.Jumlah > sumber.Stok {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Jumlah melebihi stok sumber daya",
		})
		return
	}

	s.nextPeminjamanID++
	p.ID = s.nextPeminjamanID
	p.Status = "aktif"

	s.peminjaman = append(s.peminjaman, p)

	writeJSON(w, http.StatusCreated, p)
}

// B3: Menampilkan daftar peminjaman dengan filter 
func (s *Store) handleListPeminjaman(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	sumberDayaID := query.Get("sumber_daya_id")
	tanggal := strings.TrimSpace(query.Get("tanggal"))
	status := strings.TrimSpace(query.Get("status"))

	idFilter := 0
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

	if tanggal != "" {
		t, err := time.Parse("2006-01-02", tanggal)
		if err != nil || t.Format("2006-01-02") != tanggal {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Format tanggal harus YYYY-MM-DD dan tanggal harus valid",
			})
			return
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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

	writeJSON(w, http.StatusOK, hasil)
}

// B3: Menampilkan detail peminjaman berdasarkan ID
func (s *Store) handleGetPeminjaman(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "Peminjaman tidak ditemukan",
	})
}

func main() {
	log.Println("Server berjalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", NewHandler()))
}