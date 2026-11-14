package validation

import (
	"encoding/json"
	"errors"
	"io"
	"rcon-scheduler-panel/internal/errs"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrValidation       = errs.ErrBadRequest.WithCode("VALIDATION_FAILED").WithMessage("request validation failed")
	ErrMalformedJSON    = errs.ErrBadRequest.WithCode("MALFORMED_JSON").WithMessage("request body is not valid JSON")
	ErrInvalidPathParam = errs.ErrBadRequest.WithCode("INVALID_PATH_PARAMETER").WithMessage("invalid path parameter")
)

func FromBindError(err error) *errs.AppError {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		violations := make([]errs.Violation, len(ve))
		for i, fe := range ve {
			violations[i] = errs.Violation{Field: fieldPath(fe), Code: fe.Tag(), Param: fe.Param()}
		}
		return ErrValidation.WithViolations(violations).WithErr(err)
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return ErrValidation.
			WithViolations([]errs.Violation{{Field: typeErr.Field, Code: "type", Param: typeErr.Type.String()}}).
			WithErr(err)
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return ErrMalformedJSON.WithErr(err)
	}

	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		return ErrInvalidPathParam.WithErr(err)
	}

	return errs.ErrBadRequest.WithErr(err)
}

func fieldPath(fe validator.FieldError) string {
	ns := fe.Namespace()
	if _, rest, ok := strings.Cut(ns, "."); ok {
		return rest
	}
	return fe.Field()
}
