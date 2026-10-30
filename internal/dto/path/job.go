package path

type JobID struct {
	ID int64 `binding:"required" uri:"id"`
}

type JobServerID struct {
	ServerID int64 `binding:"required" uri:"server_id"`
}