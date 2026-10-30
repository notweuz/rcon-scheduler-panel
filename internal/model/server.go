package model

import (
	"strings"
	"time"

	"github.com/notweuz/rcon-scheduler-core/server"
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

func (s Server) ToCore() server.Server { // TODO: apparently core itself should have parseType. Rewrite this after adding it to core
	serverType := server.TypeSource

	if s.Protocol != nil {
		candidate := server.Type(strings.ToLower(strings.TrimSpace(s.Protocol.Name)))

		switch candidate {
		case server.TypeSource, server.TypeWebsocket, server.TypeTelnet:
			serverType = candidate
		}
	}

	srv := server.New(
		s.Address,
		int(s.Port),
		s.Password,
		serverType,
	)

	return srv
}