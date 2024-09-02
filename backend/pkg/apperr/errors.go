package apperr

import "fmt"

type BaseError struct {
	Message string
	Err error
}

type ServerError struct {
	*BaseError
}

type InvalidParamError struct {
	*BaseError
}

type NotFoundError struct {
	*BaseError
}

func (e *BaseError) Error() string {
	if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *BaseError) Unwrap() error {
    return e.Err
}