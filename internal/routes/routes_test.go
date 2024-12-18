package routes

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "maas-memes-service/internal/app"
)

func TestSetupRoutes(t *testing.T) {
    application, _ := app.NewApp()
    SetupRoutes(application)

    req, _ := http.NewRequest("GET", "/memes", nil)
    rr := httptest.NewRecorder()

    application.Router.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
