package routes

import (
    "maas-memes-service/internal/app"
    "maas-memes-service/internal/handlers"
    "maas-memes-service/internal/middleware"
)

func SetupRoutes(a *app.App) {
    a.Router.HandleFunc(
        "/memes",
        middleware.RateLimitMiddleware(a, middleware.TokenCheckMiddleware(a, handlers.GetMemeHandler)),
    ).Methods("GET")

    a.Router.HandleFunc(
        "/balance",
        middleware.RateLimitMiddleware(a, handlers.GetBalanceHandler(a.DB)),
    ).Methods("GET")

    a.Router.HandleFunc(
        "/tokens",
        middleware.RateLimitMiddleware(a, handlers.AddTokensHandler(a.DB)),
    ).Methods("POST")
}
