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
	var desc string
	if s.Description != nil {
		desc = *s.Description
	}

	var protoName string
	if s.Protocol != nil {
		protoName = s.Protocol.Name
	}

	return ServerShort{
		ID:          s.ID,
		Name:        s.Name,
		Address:     s.Address,
		Port:        s.Port,
		Protocol:    protoName,
		Description: desc,
		CreatedAt:   s.CreatedAt,
	}
}

type ExecuteResult struct {
	Response string `json:"response"`
}
