package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "maas-memes-service/internal/models"
)

func GetBalanceHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientID := r.Header.Get("X-Client-ID")
        if clientID == "" {
            http.Error(w, "unauthorized: missing client ID", http.StatusUnauthorized)
            return
        }

        var balance models.TokenBalance
        err := db.QueryRow(
            "SELECT client_id, balance FROM token_balances WHERE client_id = ?",
            clientID,
        ).Scan(&balance.ClientID, &balance.Balance)
        if err != nil {
            if err == sql.ErrNoRows {
                balance = models.TokenBalance{ClientID: clientID, Balance: 0}
            } else {
                http.Error(w, "internal server error", http.StatusInternalServerError)
                return
            }
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(balance)
    }
}

func AddTokensHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var tb models.TokenBalance
        if err := json.NewDecoder(r.Body).Decode(&tb); err != nil {
            http.Error(w, "invalid request body", http.StatusBadRequest)
            return
        }
        if tb.ClientID == "" || tb.Balance <= 0 {
            http.Error(w, "invalid client ID or balance", http.StatusBadRequest)
            return
        }

        _, err := db.Exec(
            `INSERT INTO token_balances (client_id, balance, updated_at) 
             VALUES (?, ?, CURRENT_TIMESTAMP) 
             ON CONFLICT(client_id) DO UPDATE SET 
                 balance = balance + ?, 
                 updated_at = CURRENT_TIMESTAMP`,
            tb.ClientID, tb.Balance, tb.Balance,
        )
        if err != nil {
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "success"})
    }
}
