package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/TeluTrix/tarc/api/auth"
	"github.com/TeluTrix/tarc/api/db"
)

func main() {
	var database db.DB

	database.ConnectToDB()
	defer database.CloseConnection()
	if err := database.MigrateModels(); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	authService := auth.NewService(database.Conn)
	mux := http.NewServeMux()
	authService.RegisterRoutes(mux)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("HTTP server stopped", "error", err)
		os.Exit(1)
	}
}
