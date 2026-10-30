package path

type JobID struct {
	ID int64 `binding:"required" uri:"id"`
}

