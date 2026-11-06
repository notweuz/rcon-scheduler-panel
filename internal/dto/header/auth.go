package header

type Authorization struct {
	Token string `header:"Authorization" binding:"required"`
}
