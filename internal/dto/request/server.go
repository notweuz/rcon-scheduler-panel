package request

type CreateServer struct {
	Name       string `json:"name" binding:"required,min=3,max=50"`
	Address    string `json:"address" binding:"required,min=3,max=255"`
	Port       uint16 `json:"port" binding:"required,min=1,max=65535"`
	Password   string `json:"password" binding:"required,min=8,max=24"`
	ProtocolID int64  `json:"protocol_id" binding:"required"`
}
