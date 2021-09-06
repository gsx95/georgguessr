package pkg

import "errors"


type ErrorType int

const (
	BAD_REQUEST ErrorType = iota
	NOT_FOUND
	INTERNAL
)


type BadRequestError struct {
	error
}

func BadRequestErr(msg string) BadRequestError {
	return BadRequestError{
		error: errors.New(msg),
	}
}

type NotFoundError struct {
	error
}

func NotFoundErr(msg string) NotFoundError {
	return NotFoundError{
		error: errors.New(msg),
	}
}

type InternalError struct {
	error
}

func InternalErr(msg string) InternalError {
	return InternalError{
		error: errors.New(msg),
	}
}

func ErrType(err error) ErrorType {

	switch err.(type) {
	case BadRequestError: return BAD_REQUEST
	case NotFoundError: return NOT_FOUND
	case InternalError: return INTERNAL
	default: return INTERNAL
	}
}

func PanicBadRequest(errorMsg string) {
	panic(BadRequestError{
		error: errors.New(errorMsg),
	})
}