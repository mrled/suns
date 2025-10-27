package commands

// An error type that includes an exit code
type ExitError struct {
	Code int
	Err  error
}

// Implement the error interface
func (e *ExitError) Error() string {
	return e.Err.Error()
}
func (e *ExitError) Unwrap() error {
	return e.Err
}

func ExitWithCode(code int, err error) *ExitError {
	if err == nil {
		return nil
	}
	return &ExitError{
		Code: code,
		Err:  err,
	}
}

type UsageError struct{ error }

func (e *UsageError) Unwrap() error {
	return e.error
}
