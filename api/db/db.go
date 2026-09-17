package db

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/TeluTrix/tarc/api/role"
	"github.com/TeluTrix/tarc/api/user"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	Conn *gorm.DB
}

func (db DB) ConnectToDB() {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)

	var err error
	db.Conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func (db DB) CloseConnection() {
	database, err := db.Conn.DB()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	database.Close()
}

func (db DB) MigrateModels() {
	db.Conn.AutoMigrate(role.SchemaRole{})
	db.Conn.AutoMigrate(user.SchemaUser{})
}
