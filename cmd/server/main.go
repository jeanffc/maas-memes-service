package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "maas-memes-service/internal/app"
    "maas-memes-service/internal/routes"
)

const (
    defaultPort     = "8080"
    readTimeout     = 5 * time.Second
    writeTimeout    = 10 * time.Second
    shutdownTimeout = 30 * time.Second
)

func main() {
    application, err := app.NewApp()
    if err != nil {
        log.Fatal("Error creating application:", err)
    }
    defer application.DB.Close()

    routes.SetupRoutes(application)

    srv := &http.Server{
        Addr:         ":" + defaultPort,
        Handler:      application.Router,
        ReadTimeout:  readTimeout,
        WriteTimeout: writeTimeout,
    }

    serverErrors := make(chan error, 1)
    go func() {
        application.Logger.Printf("Server starting on port %s", defaultPort)
        serverErrors <- srv.ListenAndServe()
    }()

    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

    select {
    case err := <-serverErrors:
        application.Logger.Fatalf("Error starting server: %v", err)
    case sig := <-shutdown:
        application.Logger.Printf("Start shutdown: %v", sig)
        ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
        defer cancel()

        if err := srv.Shutdown(ctx); err != nil {
            application.Logger.Printf("Graceful shutdown did not complete in %v: %v", shutdownTimeout, err)
            if err := srv.Close(); err != nil {
                application.Logger.Fatalf("Could not stop HTTP server: %v", err)
            }
        }
    }
}
