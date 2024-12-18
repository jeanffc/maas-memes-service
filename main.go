// main.go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/time/rate"
)

// Constants
const (
    defaultPort     = "8080"
    dbPath         = "./maas.db"
    readTimeout    = 5 * time.Second
    writeTimeout   = 10 * time.Second
    shutdownTimeout = 30 * time.Second
    maxOpenConns   = 25
    maxIdleConns   = 25
    connMaxLifetime = 5 * time.Minute
    rateLimit      = 100 // requests per second
    rateBurst      = 200 // burst capacity
)

// Custom errors
var (
    ErrInvalidRequest = errors.New("invalid request parameters")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrInsufficientTokens = errors.New("insufficient tokens")
    ErrDatabaseOperation = errors.New("database operation failed")
)

// Structures
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

// App encapsulates dependencies
type App struct {
    db      *sql.DB
    router  *mux.Router
    limiter *rate.Limiter
    logger  *log.Logger
}

// NewApp creates a new application instance
func NewApp() (*App, error) {
    // Initialize logger
    logger := log.New(os.Stdout, "[MaaS] ", log.LstdFlags|log.Lshortfile)

    // Initialize database
    db, err := initDB()
    if err != nil {
        return nil, err
    }

    // Create router
    router := mux.NewRouter()

    // Create rate limiter
    limiter := rate.NewLimiter(rate.Limit(rateLimit), rateBurst)

    return &App{
        db:      db,
        router:  router,
        limiter: limiter,
        logger:  logger,
    }, nil
}

// Initialize database
func initDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Set connection pool parameters
    db.SetMaxOpenConns(maxOpenConns)
    db.SetMaxIdleConns(maxIdleConns)
    db.SetConnMaxLifetime(connMaxLifetime)

    // Create tables
    createTablesSQL := `
    CREATE TABLE IF NOT EXISTS token_balances (
        client_id TEXT PRIMARY KEY,
        balance INTEGER DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS memes (
        id TEXT PRIMARY KEY,
        url TEXT NOT NULL,
        caption TEXT NOT NULL,
        query TEXT,
        latitude REAL,
        longitude REAL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_token_balances_client_id ON token_balances(client_id);
    CREATE INDEX IF NOT EXISTS idx_memes_query ON memes(query);`

    _, err = db.Exec(createTablesSQL)
    return db, err
}

// Middleware to handle rate limiting
func (app *App) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !app.limiter.Allow() {
            app.respondWithError(w, http.StatusTooManyRequests, "rate limit exceeded")
            return
        }
        next(w, r)
    }
}

// Middleware to check token
func (app *App) checkTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        clientID := r.Header.Get("X-Client-ID")
        if clientID == "" {
            app.respondWithError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
            return
        }

        // Use transaction for token check and update
        tx, err := app.db.BeginTx(ctx, nil)
        if err != nil {
            app.logger.Printf("Error starting transaction: %v", err)
            app.respondWithError(w, http.StatusInternalServerError, ErrDatabaseOperation.Error())
            return
        }
        defer tx.Rollback()

        var balance int
        err = tx.QueryRowContext(ctx, 
            "SELECT balance FROM token_balances WHERE client_id = ?", 
            clientID).Scan(&balance)
        if err != nil || balance <= 0 {
            app.respondWithError(w, http.StatusPaymentRequired, ErrInsufficientTokens.Error())
            return
        }

        _, err = tx.ExecContext(ctx,
            "UPDATE token_balances SET balance = balance - 1, updated_at = CURRENT_TIMESTAMP WHERE client_id = ?",
            clientID)
        if err != nil {
            app.logger.Printf("Error updating balance: %v", err)
            app.respondWithError(w, http.StatusInternalServerError, ErrDatabaseOperation.Error())
            return
        }

        if err = tx.Commit(); err != nil {
            app.logger.Printf("Error committing transaction: %v", err)
            app.respondWithError(w, http.StatusInternalServerError, ErrDatabaseOperation.Error())
            return
        }

        next(w, r)
    }
}

// Handler to get memes
func (app *App) getMemeHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
    if err != nil {
        app.respondWithError(w, http.StatusBadRequest, "invalid latitude")
        return
    }
    lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
    if err != nil {
        app.respondWithError(w, http.StatusBadRequest, "invalid longitude")
        return
    }

    meme := Meme{
        ID:        strconv.FormatInt(time.Now().UnixNano(), 10),
        URL:       "https://example.com/meme.jpg",
        Caption:   "A meme about " + query,
        Query:     query,
        Lat:       lat,
        Lon:       lon,
        CreatedAt: time.Now(),
    }

    app.respondWithJSON(w, http.StatusOK, meme)
}

// Handler to check token balance
func (app *App) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    clientID := r.Header.Get("X-Client-ID")
    if clientID == "" {
        app.respondWithError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
        return
    }

    var balance TokenBalance
    err := app.db.QueryRowContext(ctx,
        "SELECT client_id, balance FROM token_balances WHERE client_id = ?",
        clientID).Scan(&balance.ClientID, &balance.Balance)
    if err != nil {
        if err == sql.ErrNoRows {
            app.respondWithJSON(w, http.StatusOK, TokenBalance{ClientID: clientID, Balance: 0})
            return
        }
        app.logger.Printf("Error getting balance: %v", err)
        app.respondWithError(w, http.StatusInternalServerError, ErrDatabaseOperation.Error())
        return
    }

    app.respondWithJSON(w, http.StatusOK, balance)
}

// Handler to add tokens
func (app *App) addTokensHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var balance TokenBalance
    if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
        app.respondWithError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if balance.ClientID == "" || balance.Balance <= 0 {
        app.respondWithError(w, http.StatusBadRequest, "invalid client ID or balance")
        return
    }

    _, err := app.db.ExecContext(ctx, `
        INSERT INTO token_balances (client_id, balance, updated_at) 
        VALUES (?, ?, CURRENT_TIMESTAMP) 
        ON CONFLICT(client_id) 
        DO UPDATE SET 
            balance = balance + ?,
            updated_at = CURRENT_TIMESTAMP`,
        balance.ClientID, balance.Balance, balance.Balance)
    
    if err != nil {
        app.logger.Printf("Error adding tokens: %v", err)
        app.respondWithError(w, http.StatusInternalServerError, ErrDatabaseOperation.Error())
        return
    }

    app.respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Helper function to respond with JSON
func (app *App) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(payload)
}

// Helper function to respond with error
func (app *App) respondWithError(w http.ResponseWriter, code int, message string) {
    app.respondWithJSON(w, code, map[string]string{"error": message})
}

// Setup routes
func (app *App) setupRoutes() {
    app.router.HandleFunc("/memes", app.rateLimitMiddleware(app.checkTokenMiddleware(app.getMemeHandler))).Methods("GET")
    app.router.HandleFunc("/balance", app.rateLimitMiddleware(app.getBalanceHandler)).Methods("GET")
    app.router.HandleFunc("/tokens", app.rateLimitMiddleware(app.addTokensHandler)).Methods("POST")
}

func main() {
    // Create new app instance
    app, err := NewApp()
    if err != nil {
        log.Fatal("Error creating application:", err)
    }
    defer app.db.Close()

    // Setup routes
    app.setupRoutes()

    // Create server with timeouts
    srv := &http.Server{
        Addr:         ":" + defaultPort,
        Handler:      app.router,
        ReadTimeout:  readTimeout,
        WriteTimeout: writeTimeout,
    }

    // Channel to listen for errors coming from the listener.
    serverErrors := make(chan error, 1)
    
    // Start the service listening for requests.
    go func() {
        app.logger.Printf("Server starting on port %s", defaultPort)
        serverErrors <- srv.ListenAndServe()
    }()

    // Channel to listen for an interrupt or terminate signal from the OS.
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

    // Blocking main and waiting for shutdown.
    select {
    case err := <-serverErrors:
        app.logger.Fatalf("Error starting server: %v", err)

    case sig := <-shutdown:
        app.logger.Printf("Start shutdown: %v", sig)
        
        // Create context for graceful shutdown.
        ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
        defer cancel()

        // Asking listener to shut down and shed load.
        if err := srv.Shutdown(ctx); err != nil {
            app.logger.Printf("Graceful shutdown did not complete in %v: %v", shutdownTimeout, err)
            if err := srv.Close(); err != nil {
                app.logger.Fatalf("Could not stop http server: %v", err)
            }
        }
    }
}