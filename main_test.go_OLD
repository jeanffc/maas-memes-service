package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "maas_test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Override dbPath (now a var instead of const in main.go)
	tempDBPath := filepath.Join(dir, "maas_test.db")
	originalDBPath := dbPath
	dbPath = tempDBPath

	code := m.Run()

	// Restore the original dbPath
	dbPath = originalDBPath
	os.Exit(code)
}

func TestAppInitialization(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.db.Close()

	if app.db == nil {
		t.Error("Expected a valid DB connection")
	}
	if app.router == nil {
		t.Error("Expected a valid router")
	}
}

func TestRoutes(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.db.Close()

	app.setupRoutes()

	routes := []struct {
		path   string
		method string
	}{
		{"/memes", http.MethodGet},
		{"/balance", http.MethodGet},
		{"/tokens", http.MethodPost},
	}

	for _, r := range routes {
		req := httptest.NewRequest(r.method, r.path, nil)
		w := httptest.NewRecorder()

		app.router.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s %s returned 404", r.method, r.path)
		}
	}
}

func TestAddTokens(t *testing.T) {
	app, _ := NewApp()
	defer app.db.Close()
	app.setupRoutes()

	tb := TokenBalance{ClientID: "client_add", Balance: 100}
	body, _ := json.Marshal(tb)

	req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var balance int
	err := app.db.QueryRow("SELECT balance FROM token_balances WHERE client_id = ?", tb.ClientID).Scan(&balance)
	if err != nil {
		t.Fatalf("DB query error: %v", err)
	}
	if balance != 100 {
		t.Errorf("Expected balance=100, got %d", balance)
	}
}

func TestGetBalance(t *testing.T) {
	app, _ := NewApp()
	defer app.db.Close()
	app.setupRoutes()

	// Seed the DB
	app.db.Exec("INSERT INTO token_balances (client_id, balance) VALUES (?, ?)", "client_balance", 50)

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req.Header.Set("X-Client-ID", "client_balance")
	w := httptest.NewRecorder()

	app.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var tb TokenBalance
	json.Unmarshal(w.Body.Bytes(), &tb)
	if tb.Balance != 50 {
		t.Errorf("Expected 50, got %d", tb.Balance)
	}
}

func TestGetMemes(t *testing.T) {
	app, _ := NewApp()
	defer app.db.Close()
	app.setupRoutes()

	app.db.Exec("INSERT INTO token_balances (client_id, balance) VALUES (?, ?)", "client_memes", 10)

	req := httptest.NewRequest(http.MethodGet, "/memes?lat=40.73061&lon=-73.935242&query=food", nil)
	req.Header.Set("X-Client-ID", "client_memes")
	w := httptest.NewRecorder()

	app.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var meme Meme
	json.Unmarshal(w.Body.Bytes(), &meme)
	if meme.ID == "" || meme.URL == "" {
		t.Error("Meme response missing required fields")
	}

	var balanceAfter int
	err := app.db.QueryRow("SELECT balance FROM token_balances WHERE client_id = ?", "client_memes").Scan(&balanceAfter)
	if err != nil {
		t.Fatalf("DB query error: %v", err)
	}
	if balanceAfter != 9 {
		t.Errorf("Expected balance=9, got %d", balanceAfter)
	}
}

