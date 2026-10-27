package model

type Job struct {
	ID         int64 `gorm:"primaryKey"`
	ServerID   int64 
	Server     *Server
	Name       string
	Cron       string
	LogResults bool

	Steps []Step `gorm:"type:jsonb;serializer:json;not null"`
}

type Step struct {
	Command string `json:"command,omitempty"`
	Wait    string `json:"wait,omitempty"`
}
