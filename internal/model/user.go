package model

import (
	"time"
)

type User struct {
	ID       int64  `gorm:"primaryKey"`
	Username string `gorm:"not null"`
	Password string `gorm:"not null"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (User) TableName() string {
	return "users"
}
