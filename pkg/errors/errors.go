package errors

import (
	goErrors "errors"

	"github.com/go-playground/validator/v10"
	"github.com/latolukasz/beeorm/v2"
)

type FieldErrors map[string]string

type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "unauthorized"
}

type NotFoundError string

func (e NotFoundError) Error() string {
	return string(e)
}

type PermissionError struct {
	Message string
}

func (e *PermissionError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "permission denied"
}

func HandleErrors(formErrors interface{}) error {
	fErrors, ok := formErrors.(validator.ValidationErrors)

	if !ok {
		return nil
	}

	var fe FieldErrors = make(map[string]string)
	for _, errors := range fErrors {
		fe[errors.Field()] = errors.Translate(nil)
	}

	return fe
}

func HandleCustomErrors(formErrors map[string]string) error {
	var fe FieldErrors = make(map[string]string)
	for field, msg := range formErrors {
		fe[field] = msg
	}

	return fe
}

func (fe FieldErrors) Error() string {
	var result string
	for _, val := range fe {
		result += val + "\n\r"
	}

	return result
}

func HandleFlushWithCheckError(err, duplicatedKeyError error) error {
	_, ok := err.(*beeorm.DuplicatedKeyError)
	if ok {
		return duplicatedKeyError
	}

	// TODO: check with Lukasz why he remove this error
	//foreignKeyErr, ok := err.(*beeorm.ForeignKeyError)
	//if ok {
	//	return foreignKeyErr
	//}

	return goErrors.New("unexpected error happened")
}
