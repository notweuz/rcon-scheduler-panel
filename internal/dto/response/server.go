package response

import (
	"rcon-scheduler-panel/internal/model"
	"time"
)

type ServerShort struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Port        uint16 `json:"port"`
	Protocol    string `json:"protocol"`
	Description string `json:"description,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

func ServerShortFromModel(s model.Server) ServerShort {
	return ServerShort{
		ID:          s.ID,
		Name:        s.Name,
		Address:     s.Address,
		Port:        s.Port,
		Protocol:    s.Protocol.Name,
		Description: *s.Description,
		CreatedAt:   s.CreatedAt,
	}
}
