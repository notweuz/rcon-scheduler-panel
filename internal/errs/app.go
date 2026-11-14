package errs

type Violation struct {
	Field string
	Code  string
	Param string
}

type AppError struct {
	Kind       Kind
	Code       string
	Message    string
	Err        error
	Violations []Violation

	base *AppError
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
	if e.Kind == t.Kind && e.Code == t.Code {
		return true
	}
	return e.base != nil && e.base == t
}

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) clone() *AppError {
	cp := *e
	if cp.base == nil {
		cp.base = e
	}
	return &cp
}

func (e *AppError) WithCode(code string) *AppError {
	cp := e.clone()
	cp.Code = code
	return cp
}

func (e *AppError) WithMessage(msg string) *AppError {
	cp := e.clone()
	cp.Message = msg
	return cp
}

func (e *AppError) WithErr(err error) *AppError {
	cp := e.clone()
	cp.Err = err
	return cp
}

func (e *AppError) WithViolations(v []Violation) *AppError {
	cp := e.clone()
	cp.Violations = v
	return cp
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
