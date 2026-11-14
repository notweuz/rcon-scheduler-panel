package path

type JobID struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type JobServerID struct {
	ServerID int64 `uri:"id" binding:"required,min=1"`
}
