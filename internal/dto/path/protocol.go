package path

type ProtocolID struct {
	ID int64 `uri:"protocol" binding:"required"`
}
