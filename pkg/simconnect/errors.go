package simconnect

import "fmt"

// SimConnectError represents a SimConnect-specific error
type SimConnectError struct {
	Function string
	HRESULT  uint32
	Message  string
}

func (e *SimConnectError) Error() string {
	return fmt.Sprintf("SimConnect %s failed: %s (HRESULT: 0x%08X)", e.Function, e.Message, e.HRESULT)
}

// NewSimConnectError creates a new SimConnect error
func NewSimConnectError(function string, hresult uint32, message string) *SimConnectError {
	return &SimConnectError{
		Function: function,
		HRESULT:  hresult,
		Message:  message,
	}
}

// IsHRESULTSuccess checks if an HRESULT indicates success
func IsHRESULTSuccess(hresult uint32) bool {
	return hresult == S_OK
}

// GetHRESULTMessage returns a human-readable message for common HRESULT values
func GetHRESULTMessage(hresult uint32) string {
	switch hresult {
	case S_OK:
		return "Success"
	case E_FAIL:
		return "General failure"
	case E_INVALIDARG:
		return "Invalid argument"
	case STATUS_REMOTE_DISCONNECT:
		return "Remote connection lost"
	default:
		return fmt.Sprintf("Unknown error (0x%08X)", hresult)
	}
}
