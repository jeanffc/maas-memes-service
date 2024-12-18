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
    rateLimit = 100
    rateBurst = 200
)

type App struct {
    DB      *sql.DB
    Router  *mux.Router
    Limiter *rate.Limiter
    Logger  *log.Logger
}

func NewApp() (*App, error) {
    logger := log.New(os.Stdout, "[MaaS] ", log.LstdFlags|log.Lshortfile)

    database, err := db.InitDB()
    if err != nil {
        return nil, err
    }

    router := mux.NewRouter()
    limiter := rate.NewLimiter(rate.Limit(rateLimit), rateBurst)

    return &App{
        DB:      database,
        Router:  router,
        Limiter: limiter,
        Logger:  logger,
    }, nil
}
