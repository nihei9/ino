package multierr

import (
	"fmt"
	"strings"
)

type MultipleErrors struct {
	errs []error
}

func Bundle(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	return &MultipleErrors{
		errs: errs,
	}
}

func (e *MultipleErrors) Error() string {
	if len(e.errs) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprint(&b, e.errs[0].Error())
	for _, err := range e.errs[1:] {
		fmt.Fprintf(&b, "\n\n%v", err)
	}
	return b.String()
}
