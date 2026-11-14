package response

type JobList struct {
	Items []Job `json:"items"`
}

type ServerList struct {
	Items []ServerShort `json:"items"`
}

type ProtocolList struct {
	Items []Protocol `json:"items"`
}
