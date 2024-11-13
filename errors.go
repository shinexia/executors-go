package executors

import (
	"fmt"
	"io"
	"log"
	"strings"
)

// errorList should not be used directly, see TestErrors
type errorList []error

func (es errorList) Error() string {
	if len(es) <= 1 {
		log.Panic("should not come here:", es)
	}
	b := strings.Builder{}
	b.Write([]byte("["))
	for i, e := range es {
		if i > 0 {
			b.Write([]byte(","))
		}
		b.WriteString(e.Error())
	}
	b.Write([]byte("]"))
	return b.String()
}

func (es errorList) Format(s fmt.State, verb rune) {
	n := len(es) - 1
	switch verb {
	case 'v':
		if s.Flag('+') {
			for i, w := range es {
				if i < n {
					fmt.Fprintf(s, "%+v\n", w)
				} else {
					fmt.Fprintf(s, "%+v", w)
				}
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, es.Error())
	case 'q':
		fmt.Fprintf(s, "%q", es.Error())
	}
}

func AppendError(err error, tail ...error) error {
	var errOut errorList
	if err == nil {
		if len(tail) == 1 {
			return tail[0]
		}
		errOut = make(errorList, 0, len(tail))
	} else if elist, ok := err.(errorList); ok {
		errOut = elist
	} else {
		errOut = make(errorList, 1, 1+len(tail))
		errOut[0] = err
	}
	for _, e := range tail {
		if e != nil {
			errOut = append(errOut, e)
		}
	}
	switch len(errOut) {
	case 0:
		return nil
	case 1:
		return errOut[0]
	default:
		return errOut
	}
}

// Runtime error, cannot be handled by retrying
type runtimeError struct {
	Err error
}

func NewRuntimeError(err error) error {
	return runtimeError{
		Err: err,
	}
}

func NewRuntimeErrorf(format string, args ...any) error {
	return runtimeError{
		Err: fmt.Errorf(format, args...),
	}
}

func (e runtimeError) Error() string {
	return e.Err.Error()
}

func (e runtimeError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", e.Err)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Err.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Err.Error())
	}
}

func IsRuntimeError(err error) bool {
	e, ok := err.(errorList)
	if ok {
		for _, x := range e {
			if IsRuntimeError(x) {
				return true
			}
		}
	}
	_, ok = err.(runtimeError)
	return ok
}
