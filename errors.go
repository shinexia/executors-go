package executors

import (
	"fmt"
	"io"
	"log"
	"strings"
)

type errorList []error

// unwrapError the same to errors.joinError.Unwrap
type unwrapError interface {
	Unwrap() []error
}

func (es errorList) Error() string {
	if len(es) <= 1 {
		log.Panic("[SHOULD_NEVER_HAPPEN]:", es)
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

// Unwrap implements the unwrapError interface
func (es errorList) Unwrap() []error {
	return []error(es)
}

// Format implements the fmt.Formatter interface
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

func AppendError(errIn ...error) error {
	switch len(errIn) {
	case 0:
		return nil
	case 1:
		return errIn[0]
	case 2:
		if errIn[0] == nil {
			return errIn[1]
		}
		if errIn[1] == nil {
			return errIn[0]
		}
	}
	var (
		errOut errorList
		err    = errIn[0]
		tail   = errIn[1:]
	)
	switch we := err.(type) {
	case nil:
		if len(tail) == 1 {
			return tail[0]
		}
		errOut = make(errorList, 0, len(tail))
	case errorList:
		errOut = we
	case unwrapError:
		errOut = errorList(we.Unwrap())
	default:
		errOut = make(errorList, 1, 1+len(tail))
		errOut[0] = err
	}
	for _, e := range tail {
		switch v := e.(type) {
		case nil:
			continue
		case errorList:
			errOut = append(errOut, v...)
		case unwrapError:
			errOut = append(errOut, v.Unwrap()...)
		default:
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
	_, ok := err.(runtimeError)
	if ok {
		return true
	}
	var es []error
	switch e := err.(type) {
	case errorList:
		es = e
	case unwrapError:
		es = e.Unwrap()
	default:
		return false
	}
	for _, x := range es {
		if IsRuntimeError(x) {
			return true
		}
	}
	return false
}
