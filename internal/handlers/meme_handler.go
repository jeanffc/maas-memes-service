package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "maas-memes-service/internal/models"
)

func GetMemeHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
    if err != nil {
        http.Error(w, "invalid latitude", http.StatusBadRequest)
        return
    }
    lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
    if err != nil {
        http.Error(w, "invalid longitude", http.StatusBadRequest)
        return
    }

    meme := models.Meme{
        ID:        strconv.FormatInt(time.Now().UnixNano(), 10),
        URL:       "https://example.com/meme.jpg",
        Caption:   "A meme about " + query,
        Query:     query,
        Lat:       lat,
        Lon:       lon,
        CreatedAt: time.Now(),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(meme)
}
