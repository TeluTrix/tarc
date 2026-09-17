package main

import (
	"log/slog"
	"os"

	"github.com/TeluTrix/tarc/api/db"
)

func init() {
	var database db.DB

	database.ConnectToDB()
	if err := database.MigrateModels(); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	database.CloseConnection()
}

func main() {}
