package request

type CreateProtocol struct {
	Name        string  `json:"name" binding:"required,min=3,max=50"`
	Description *string `json:"description"`
}
