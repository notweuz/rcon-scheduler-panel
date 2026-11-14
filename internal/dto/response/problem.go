package response

const ProblemContentType = "application/problem+json"

type Problem struct {
	Title    string         `json:"title" example:"Not Found"`
	Status   int            `json:"status" example:"404"`
	Detail   string         `json:"detail,omitempty" example:"server not found"`
	Instance string         `json:"instance,omitempty" example:"/api/servers/42"`
	Code     string         `json:"code" example:"SERVER_NOT_FOUND"`
	Errors   []InvalidParam `json:"errors,omitempty"`
}

type InvalidParam struct {
	Field string `json:"field" example:"name"`
	Code  string `json:"code" example:"min"`
	Param string `json:"param,omitempty" example:"3"`
}
