package middleware

import (
    "net/http"
    "maas-memes-service/internal/app"
)

func TokenCheckMiddleware(a *app.App, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientID := r.Header.Get("X-Client-ID")
        if clientID == "" {
            http.Error(w, "unauthorized: missing client ID", http.StatusUnauthorized)
            return
        }

        var balance int
        err := a.DB.QueryRow("SELECT balance FROM token_balances WHERE client_id = ?", clientID).Scan(&balance)
        if err != nil || balance <= 0 {
            http.Error(w, "payment required: insufficient tokens", http.StatusPaymentRequired)
            return
        }

        _, err = a.DB.Exec(
            "UPDATE token_balances SET balance = balance - 1, updated_at = CURRENT_TIMESTAMP WHERE client_id = ?",
            clientID,
        )
        if err != nil {
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }

        next(w, r)
    }
}
