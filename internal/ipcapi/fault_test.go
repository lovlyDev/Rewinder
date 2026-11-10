package ipcapi

import (
	"errors"
	"fmt"
	"testing"
)

func TestFaultContract(t *testing.T) {
	cause := errors.New("native failure")
	fault := &Fault{Code: "restore.launch", Params: map[string]any{"app": "test"}, Cause: cause}
	wrapped := fmt.Errorf("operation: %w", fault)
	code, params := ErrorCode(wrapped)
	if code != "restore.launch" || params["app"] != "test" || !errors.Is(wrapped, cause) {
		t.Fatal(code, params, wrapped)
	}
	code, params = ErrorCode(cause)
	if code != "restore.failed" || params != nil {
		t.Fatal(code, params)
	}
}
