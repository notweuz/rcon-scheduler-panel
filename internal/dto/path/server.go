package path

type ServerID struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}
