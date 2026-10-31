package path

type ServerID struct {
	ID int64 `binding:"required" uri:"id"`
}
