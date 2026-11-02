package response

import "rcon-scheduler-panel/internal/model"

type Protocol struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func ProtocolFromModel(p model.Protocol) Protocol {
	var desc string
	if p.Description != nil {
		desc = *p.Description
	}

	return Protocol{
		ID:          p.ID,
		Name:        p.Name,
		Description: desc,
	}
}
