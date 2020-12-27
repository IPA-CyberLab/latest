package fetch

import "fmt"

type ErrSoftwareIdParseFailed struct {
	Input       string
	HandlerName string
	Err         error
}

func (e ErrSoftwareIdParseFailed) Error() string {
	return fmt.Sprintf("%s handler has failed to process softwareId %q", e.HandlerName, e.Input)
}

func (e ErrSoftwareIdParseFailed) Unwrap() error {
	return e.Err
}
