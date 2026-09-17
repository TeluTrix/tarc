package role

import "github.com/google/uuid"

type SchemaRole struct {
	ID   uuid.UUID `gorm:"primaryKey"`
	Name string
}
