package errs

type Kind uint8

const (
	Internal Kind = iota
	Unauthenticated
	MethodNotAllowed
	BadRequest
	Forbidden
	NotFound
	Conflict
	BadGateway
)