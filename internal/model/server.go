package model

import (
	"time"
)

type Server struct {
	ID          int64  `gorm:"primaryKey"`
	Name        string
	Address     string
	Port 		uint16
	Password    string
	ProtocolID  int64
	Protocol    *Protocol
	Description *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Server) TableName() string {
	return "servers"
}