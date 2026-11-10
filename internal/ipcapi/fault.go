package ipcapi

import "errors"

// Fault is the stable UI contract. Native details stay in the log via Cause.
type Fault struct {
	Code   string
	Params map[string]any
	Cause  error
}

func (f *Fault) Error() string {
	if f.Cause != nil {
		return f.Code + ": " + f.Cause.Error()
	}
	return f.Code
}
func (f *Fault) Unwrap() error { return f.Cause }
func ErrorCode(err error) (string, map[string]any) {
	var fault *Fault
	if errors.As(err, &fault) {
		return fault.Code, fault.Params
	}
	return "restore.failed", nil
}
