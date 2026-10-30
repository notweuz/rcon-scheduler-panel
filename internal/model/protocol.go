package model

type Protocol struct {
	ID          int64  `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description *string
}

func (Protocol) TableName() string {
	return "protocols"
}
