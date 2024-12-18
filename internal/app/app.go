package app

import (
    "log"
    "os"
    "database/sql"
    "github.com/gorilla/mux"
    "golang.org/x/time/rate"
    "maas-memes-service/internal/db"
)

const (
    dbPath         = "./maas.db"
    rateLimit      = 100
    rateBurst      = 200
)

type App struct {
    DB      *sql.DB
    Router  *mux.Router
    Limiter *rate.Limiter
    Logger  *log.Logger
}

func NewApp() (*App, error) {
    logger := log.New(os.Stdout, "[MaaS] ", log.LstdFlags|log.Lshortfile)
    database, err := db.InitDB(dbPath) 
    if err != nil {
        return nil, err
    }

    limiter := rate.NewLimiter(rate.Limit(rateLimit), rateBurst)
    router := mux.NewRouter()

    return &App{
        DB:      database,
        Router:  router,
        Limiter: limiter,
        Logger:  logger,
    }, nil
}
