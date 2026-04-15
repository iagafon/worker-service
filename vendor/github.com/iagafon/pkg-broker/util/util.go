package util

import "errors"

func Coalesce(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

type notCriticalError struct {
	err error
}

func (e *notCriticalError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *notCriticalError) Unwrap() error {
	return e.err
}

// NotCriticalError оборачивает ошибку как некритичную.
func NotCriticalError(err error) error {
	if err == nil {
		return nil
	}
	return &notCriticalError{err: err}
}

// IsNotCriticalError проверяет, является ли ошибка некритичной.
func IsNotCriticalError(err error) bool {
	var target *notCriticalError
	return errors.As(err, &target)
}
