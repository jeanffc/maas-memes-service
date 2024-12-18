package models

import "time"

type Meme struct {
    ID        string    `json:"id"`
    URL       string    `json:"url"`
    Caption   string    `json:"caption"`
    Query     string    `json:"query"`
    Lat       float64   `json:"latitude"`
    Lon       float64   `json:"longitude"`
    CreatedAt time.Time `json:"created_at"`
}

type TokenBalance struct {
    ClientID string `json:"client_id"`
    Balance  int    `json:"balance"`
}
