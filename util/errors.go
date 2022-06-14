package util

import "fmt"

func Error(funcName string, reason string) error {
	return fmt.Errorf("%s: %s", funcName, reason)
}
func Errorf(funcName string, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %s", funcName, fmt.Sprintf(format, args...))
}
func Wrap(funcName string, reason string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %s\n%w", funcName, reason, err)
	}
	return nil
}
