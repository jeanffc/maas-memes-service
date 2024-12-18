package middleware

import (
    "net/http"

    "maas-memes-service/internal/app"
)

func RateLimitMiddleware(a *app.App, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !a.Limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next(w, r)
    }
}
