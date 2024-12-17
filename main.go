package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
)

// Structures
type Meme struct {
    ID      string  `json:"id"`
    URL     string  `json:"url"`
    Caption string  `json:"caption"`
    Query   string  `json:"query"`
    Lat     float64 `json:"latitude"`
    Lon     float64 `json:"longitude"`
}

type TokenBalance struct {
    ClientID string `json:"client_id"`
    Balance  int    `json:"balance"`
}

// Global database connection
var db *sql.DB

// Initialize database
func initDB() error {
    var err error
    db, err = sql.Open("sqlite3", "./maas.db")
    if err != nil {
        return err
    }

    // Create tables
    createTablesSQL := `
    CREATE TABLE IF NOT EXISTS token_balances (
        client_id TEXT PRIMARY KEY,
        balance INTEGER DEFAULT 0
    );
    CREATE TABLE IF NOT EXISTS memes (
        id TEXT PRIMARY KEY,
        url TEXT,
        caption TEXT,
        query TEXT
    );`

    _, err = db.Exec(createTablesSQL)
    return err
}

// Middleware to check token
func checkToken(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientID := r.Header.Get("X-Client-ID")
        if clientID == "" {
            http.Error(w, "Missing client ID", http.StatusUnauthorized)
            return
        }

        // Check token balance
        var balance int
        err := db.QueryRow("SELECT balance FROM token_balances WHERE client_id = ?", clientID).Scan(&balance)
        if err != nil || balance <= 0 {
            http.Error(w, "Insufficient tokens", http.StatusPaymentRequired)
            return
        }

        // Deduct token
        _, err = db.Exec("UPDATE token_balances SET balance = balance - 1 WHERE client_id = ?", clientID)
        if err != nil {
            http.Error(w, "Error processing token", http.StatusInternalServerError)
            return
        }

        next(w, r)
    }
}

// Handler to get memes
func getMemeHandler(w http.ResponseWriter, r *http.Request) {
    // Get query parameters
    query := r.URL.Query().Get("query")
    lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
    lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)

    // Generate a simple meme (in real world, this would be more sophisticated)
    meme := Meme{
        ID:      strconv.FormatInt(time.Now().Unix(), 10),
        URL:     "https://example.com/meme.jpg",
        Caption: "A meme about " + query,
        Query:   query,
        Lat:     lat,
        Lon:     lon,
    }

    json.NewEncoder(w).Encode(meme)
}

// Handler to check token balance
func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
    clientID := r.Header.Get("X-Client-ID")
    if clientID == "" {
        http.Error(w, "Missing client ID", http.StatusUnauthorized)
        return
    }

    var balance TokenBalance
    err := db.QueryRow("SELECT client_id, balance FROM token_balances WHERE client_id = ?", clientID).
        Scan(&balance.ClientID, &balance.Balance)
    if err != nil {
        http.Error(w, "Error getting balance", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(balance)
}

// Handler to add tokens
func addTokensHandler(w http.ResponseWriter, r *http.Request) {
    var balance TokenBalance
    if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Insert or update token balance
    _, err := db.Exec(`
        INSERT INTO token_balances (client_id, balance) 
        VALUES (?, ?) 
        ON CONFLICT(client_id) 
        DO UPDATE SET balance = balance + ?`,
        balance.ClientID, balance.Balance, balance.Balance)
    
    if err != nil {
        http.Error(w, "Error adding tokens", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func main() {
    // Initialize database
    if err := initDB(); err != nil {
        log.Fatal("Error initializing database:", err)
    }
    defer db.Close()

    // Create router
    r := mux.NewRouter()

    // Define routes
    r.HandleFunc("/memes", checkToken(getMemeHandler)).Methods("GET")
    r.HandleFunc("/balance", getBalanceHandler).Methods("GET")
    r.HandleFunc("/tokens", addTokensHandler).Methods("POST")

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}