package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestGetMemeHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/memes?lat=40.73061&lon=-73.935242&query=funny", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(GetMemeHandler)

    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    response := rr.Body.String()
    assert.Contains(t, response, "A meme about funny")
}
