package role

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AdminRole = "ADMIN"
	UserRole  = "USER"
)

type SchemaRole struct {
	Name string `gorm:"type:varchar(191);primaryKey"`
}

func SeedDefaults(db *gorm.DB) error {
	roles := []SchemaRole{
		{Name: AdminRole},
		{Name: UserRole},
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&roles).Error
}
