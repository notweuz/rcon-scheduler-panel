package response

type List[T any] struct {
	Items []T `json:"items"`
}

func MapList[M, T any](models []M, fn func(M) T) List[T] {
	items := make([]T, len(models))
	for i, m := range models {
		items[i] = fn(m)
	}
	return List[T]{Items: items}
}
