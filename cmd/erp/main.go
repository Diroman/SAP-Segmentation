package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Row struct {
	AddressSapID string `json:"address_sap_id"`
	AdrSegment   string `json:"adr_segment"`
	SegmentID    int64  `json:"segment_id"`
}

var mockData = generateMockData(150)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	runServer(ctx)
}

func runServer(ctx context.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/ords/bsm/segmentation/get_segmentation", getSegmentation).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server on :8080")
	log.Fatal(srv.ListenAndServe())
}

func getSegmentation(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Basic 4Dfddf5:jKlljHGH" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("p_limit")
	offsetStr := query.Get("p_offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	total := len(mockData)
	start := offset
	end := int(math.Min(float64(start+limit), float64(total)))

	if start >= total {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Row{})
		return
	}

	data := mockData[start:end]

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func generateMockData(count int) []*Row {
	data := make([]*Row, count)
	for i := 0; i < count; i++ {
		data[i] = &Row{
			AddressSapID: "SAP_" + strconv.Itoa(i+1),
			AdrSegment:   "SEG_" + string(rune('A'+(i%5))), // A, B, C, D, E
			SegmentID:    int64(i + 1),
		}
	}
	return data
}
