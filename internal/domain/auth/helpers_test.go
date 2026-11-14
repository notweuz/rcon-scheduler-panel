package auth_test

import "rcon-scheduler-panel/internal/dto/response"

func fieldNames(params []response.InvalidParam) []string {
	names := make([]string, len(params))
	for i, p := range params {
		names[i] = p.Field
	}
	return names
}

func fieldCode(params []response.InvalidParam, field string) string {
	for _, p := range params {
		if p.Field == field {
			return p.Code
		}
	}
	return ""
}
