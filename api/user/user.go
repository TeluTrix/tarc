package user

import (
	"time"

	"github.com/TeluTrix/tarc/api/role"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchemaUser struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey"`
	Email        string    `gorm:"type:varchar(191);uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"column:password;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt  `gorm:"index"`
	RoleName     string          `gorm:"type:varchar(191);not null"`
	Role         role.SchemaRole `gorm:"foreignKey:RoleName;references:Name"`
}
