package errs

import (
	"fmt"
	"runtime"
)

func Error(reason string) error {
	counter, _, _, _ := runtime.Caller(1)
	return fmt.Errorf("%s: %s", runtime.FuncForPC(counter).Name(), reason)
}
func Errorf(format string, args ...interface{}) error {
	counter, _, _, _ := runtime.Caller(1)
	return fmt.Errorf("%s: %s", runtime.FuncForPC(counter).Name(), fmt.Sprintf(format, args...))
}
func Wrap(reason string, err error) error {
	counter, _, _, _ := runtime.Caller(1)
	return fmt.Errorf("%s: %s\n%w", runtime.FuncForPC(counter).Name(), reason, err)
}

func WrapRecoveredPanic(p interface{}, stack []byte) error {
	counter, _, _, _ := runtime.Caller(1)
	if err, ok := p.(error); ok {
		return fmt.Errorf("%s: recovered from panic at %s\n%w", runtime.FuncForPC(counter).Name(), string(stack), err)
	}
	return fmt.Errorf("%s: recovered from panic at %s\n%v", runtime.FuncForPC(counter).Name(), string(stack), p)
}
