package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "maas-memes-service/internal/db"
    _ "github.com/mattn/go-sqlite3"
)

func setupTestDB() (*sql.DB, error) {
    return db.InitDB(":memory:")
}

func TestGetBalanceHandler(t *testing.T) {
    database, err := setupTestDB()
    assert.NoError(t, err)
    defer database.Close()

    _, err = database.Exec(`INSERT INTO token_balances (client_id, balance) VALUES ('test-client', 100)`)
    assert.NoError(t, err)

    req, _ := http.NewRequest("GET", "/balance", nil)
    req.Header.Set("X-Client-ID", "test-client")

    rr := httptest.NewRecorder()
    handler := GetBalanceHandler(database)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    var response map[string]interface{}
    err = json.Unmarshal(rr.Body.Bytes(), &response)
    assert.NoError(t, err)

    assert.Equal(t, "test-client", response["client_id"])
    assert.Equal(t, float64(100), response["balance"])
}
