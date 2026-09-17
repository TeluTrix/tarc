package db

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/TeluTrix/tarc/api/auth"
	"github.com/TeluTrix/tarc/api/role"
	"github.com/TeluTrix/tarc/api/user"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	Conn *gorm.DB
}

func (db *DB) ConnectToDB() {
	_ = godotenv.Load()

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)

	var err error
	db.Conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func (db *DB) CloseConnection() {
	database, err := db.Conn.DB()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	database.Close()
}

func (db *DB) MigrateModels() error {
	if err := db.migrateLegacyRoleKey(); err != nil {
		return err
	}

	if err := db.Conn.AutoMigrate(&role.SchemaRole{}, &user.SchemaUser{}, &auth.Session{}); err != nil {
		return err
	}

	return role.SeedDefaults(db.Conn)
}

func (db *DB) migrateLegacyRoleKey() error {
	if !db.Conn.Migrator().HasColumn(&role.SchemaRole{}, "ID") ||
		!db.Conn.Migrator().HasColumn(&user.SchemaUser{}, "RoleID") {
		return nil
	}

	if !db.Conn.Migrator().HasColumn(&user.SchemaUser{}, "RoleName") {
		if err := db.Conn.Exec("ALTER TABLE `schema_users` ADD COLUMN `role_name` varchar(191)").Error; err != nil {
			return err
		}
	}

	if err := db.Conn.Exec("UPDATE `schema_users` u JOIN `schema_roles` r ON u.`role_id` = r.`id` SET u.`role_name` = r.`name`").Error; err != nil {
		return err
	}

	constraints, err := db.Conn.Raw(`
		SELECT DISTINCT CONSTRAINT_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'schema_users'
		  AND COLUMN_NAME = 'role_id'
		  AND REFERENCED_TABLE_NAME = 'schema_roles'`).Rows()
	if err != nil {
		return err
	}
	defer constraints.Close()

	for constraints.Next() {
		var constraintName string
		if err := constraints.Scan(&constraintName); err != nil {
			return err
		}
		if err := db.Conn.Exec(fmt.Sprintf("ALTER TABLE `schema_users` DROP FOREIGN KEY `%s`", constraintName)).Error; err != nil {
			return err
		}
	}
	if err := constraints.Err(); err != nil {
		return err
	}

	if err := db.Conn.Exec("ALTER TABLE `schema_users` DROP COLUMN `role_id`, MODIFY COLUMN `role_name` varchar(191) NOT NULL").Error; err != nil {
		return err
	}

	return db.Conn.Exec("ALTER TABLE `schema_roles` MODIFY COLUMN `name` varchar(191) NOT NULL, DROP PRIMARY KEY, DROP COLUMN `id`, ADD PRIMARY KEY (`name`)").Error
}
