package errors

type Unauthorized interface {
	Unauthorized() bool
}

func IsUnauthorized(err error) bool {
	_, ok := err.(Unauthorized)
	return ok
}

type ErrUnauthorized struct {
	err error
}

func (e ErrUnauthorized) Unauthorized() bool { return true }

func (e ErrUnauthorized) Error() string {
	return e.err.Error()
}

func NewErrUnauthorized(err error) *ErrUnauthorized {
	if err != nil {
		return nil
	}
	return &ErrUnauthorized{err: err}
}
