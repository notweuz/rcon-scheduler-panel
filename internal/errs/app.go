package errs

type FieldError struct {
	Code  string `json:"code"`
	Param string `json:"param,omitempty"`
}

type AppError struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
	Fields  map[string]FieldError
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Kind == t.Kind && e.Code == t.Code
}

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) WithMessage(msg string) *AppError {
	cp := *e
	cp.Message = msg
	return &cp
}

func (e *AppError) WithErr(err error) *AppError {
	cp := *e
	cp.Err = err
	return &cp
}

func (e *AppError) WithFields(f map[string]FieldError) *AppError {
	cp := *e
	cp.Fields = f
	return &cp
}

func New(kind Kind, code, message string) *AppError {
	return &AppError{Kind: kind, Code: code, Message: message}
}

var (
	ErrInternalServer   = New(Internal, "INTERNAL_SERVER_ERROR", "internal server error")
	ErrNotFound         = New(NotFound, "NOT_FOUND", "resource not found")
	ErrUnauthorized     = New(Unauthenticated, "UNAUTHORIZED", "authentication required")
	ErrForbidden        = New(Forbidden, "FORBIDDEN", "access denied")
	ErrBadRequest       = New(BadRequest, "BAD_REQUEST", "invalid request")
	ErrMethodNotAllowed = New(MethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	ErrConflict         = New(Conflict, "CONFLICT", "resource conflict")
	ErrBadGateway       = New(BadGateway, "BAD_GATEWAY", "game server is unavailable")
)
