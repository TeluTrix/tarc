package main

import "github.com/TeluTrix/tarc/api/db"

func init() {
	var database db.DB

	database.ConnectToDB()
	database.MigrateModels()
	database.CloseConnection()
}

func main() {}
