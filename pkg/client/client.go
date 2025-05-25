package client

import (
	"fmt"
	"syscall"
	"unsafe"
)

// HRESULT constants
const (
	S_OK                              = 0x00000000
	E_FAIL                            = uint32(0x80004005)
	E_INVALIDARG                      = uint32(0x80070057)
	STATUS_REMOTE_DISCONNECT          = uint32(0xC000013C)
	SIMCONNECT_OPEN_CONFIGINDEX_LOCAL = 0
)

// Request ID type for SimConnect operations
type DataRequestID uint32

// Client represents a SimConnect client instance
type Client struct {
	handle uintptr          // HANDLE to SimConnect object
	dll    *syscall.LazyDLL // Reference to SimConnect.dll
	isOpen bool             // Connection state
	name   string           // Client name
}

// NewClient creates a new SimConnect client instance
func NewClient(name string) *Client {
	return &Client{
		name: name,
		dll:  syscall.NewLazyDLL("SimConnect.dll"),
	}
}

// NewClientWithDLLPath creates a new SimConnect client instance with custom DLL path
func NewClientWithDLLPath(name, dllPath string) *Client {
	return &Client{
		name: name,
		dll:  syscall.NewLazyDLL(dllPath),
	}
}

// Open establishes a connection to the SimConnect server
// Implements SimConnect_Open function
func (c *Client) Open() error {
	if c.isOpen {
		return fmt.Errorf("client is already open")
	}

	// Get the SimConnect_Open function from DLL
	proc := c.dll.NewProc("SimConnect_Open")

	// Convert name to null-terminated byte array
	nameBytes, err := syscall.BytePtrFromString(c.name)
	if err != nil {
		return fmt.Errorf("failed to convert name to bytes: %v", err)
	}
	// Call SimConnect_Open
	// HRESULT SimConnect_Open(HANDLE* phSimConnect, LPCSTR szName, HWND hWnd,
	//                         DWORD UserEventWin32, HANDLE hEventHandle, DWORD ConfigIndex)
	r1, _, _ := proc.Call(
		uintptr(unsafe.Pointer(&c.handle)), // phSimConnect
		uintptr(unsafe.Pointer(nameBytes)), // szName
		0,                                  // hWnd (NULL)
		0,                                  // UserEventWin32
		0,                                  // hEventHandle
		uintptr(SIMCONNECT_OPEN_CONFIGINDEX_LOCAL), // ConfigIndex
	)

	hresult := uint32(r1)
	if !IsHRESULTSuccess(hresult) {
		return NewSimConnectError("SimConnect_Open", hresult, GetHRESULTMessage(hresult))
	}

	c.isOpen = true
	return nil
}

// Close terminates the connection to the SimConnect server
// Implements SimConnect_Close function
func (c *Client) Close() error {
	if !c.isOpen {
		return fmt.Errorf("client is not open")
	}

	// Get the SimConnect_Close function from DLL
	proc := c.dll.NewProc("SimConnect_Close")
	// Call SimConnect_Close
	// HRESULT SimConnect_Close(HANDLE hSimConnect)
	r1, _, _ := proc.Call(c.handle)

	hresult := uint32(r1)
	if !IsHRESULTSuccess(hresult) {
		return NewSimConnectError("SimConnect_Close", hresult, GetHRESULTMessage(hresult))
	}

	c.isOpen = false
	c.handle = 0
	return nil
}

// RequestSystemState requests information from Microsoft Flight Simulator system components
// Implements SimConnect_RequestSystemState function
func (c *Client) RequestSystemState(requestID DataRequestID, state string) error {
	if !c.isOpen {
		return fmt.Errorf("client is not open")
	}

	// Get the SimConnect_RequestSystemState function from DLL
	proc := c.dll.NewProc("SimConnect_RequestSystemState")

	// Convert state string to null-terminated byte array
	stateBytes, err := syscall.BytePtrFromString(state)
	if err != nil {
		return fmt.Errorf("failed to convert state to bytes: %v", err)
	} // Call SimConnect_RequestSystemState
	// HRESULT SimConnect_RequestSystemState(HANDLE hSimConnect, SIMCONNECT_DATA_REQUEST_ID RequestID, const char* szState)
	r1, _, _ := proc.Call(
		c.handle,                            // hSimConnect
		uintptr(requestID),                  // RequestID
		uintptr(unsafe.Pointer(stateBytes)), // szState
	)

	hresult := uint32(r1)
	if !IsHRESULTSuccess(hresult) {
		return NewSimConnectError("SimConnect_RequestSystemState", hresult, GetHRESULTMessage(hresult))
	}

	return nil
}

// IsOpen returns whether the client connection is open
func (c *Client) IsOpen() bool {
	return c.isOpen
}

// GetHandle returns the internal SimConnect handle (for advanced use cases)
func (c *Client) GetHandle() uintptr {
	return c.handle
}

// GetName returns the client name
func (c *Client) GetName() string {
	return c.name
}

// SendDebugMessage sends a debug message to the Windows debug console
// Note: SimConnect does not have a built-in function to send messages to the MSFS console.
// This function uses OutputDebugString which sends messages to the Windows debug console
// that can be viewed with tools like DebugView or Visual Studio Output window.
func (c *Client) SendDebugMessage(message string) error {
	if !c.isOpen {
		return fmt.Errorf("client is not open")
	}

	// Get the OutputDebugStringA function from kernel32.dll
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	outputDebugStringA := kernel32.NewProc("OutputDebugStringA")

	// Convert message string to null-terminated byte array
	messageBytes, err := syscall.BytePtrFromString(fmt.Sprintf("[SimConnect:%s] %s", c.name, message))
	if err != nil {
		return fmt.Errorf("failed to convert message to bytes: %v", err)
	}

	// Call OutputDebugStringA
	// void OutputDebugStringA(LPCSTR lpOutputString)
	outputDebugStringA.Call(uintptr(unsafe.Pointer(messageBytes)))
	return nil
}

// AddToDataDefinition adds a simulation variable to a data definition
// Implements SimConnect_AddToDataDefinition function
func (c *Client) AddToDataDefinition(defineID DataDefinitionID, datumName, unitsName string, datumType SIMCONNECT_DATATYPE) error {
	if !c.isOpen {
		return fmt.Errorf("client is not open")
	}

	// Get the SimConnect_AddToDataDefinition function from DLL
	proc := c.dll.NewProc("SimConnect_AddToDataDefinition")

	// Convert strings to null-terminated byte arrays
	datumNameBytes, err := syscall.BytePtrFromString(datumName)
	if err != nil {
		return fmt.Errorf("failed to convert datum name to bytes: %v", err)
	}

	unitsNameBytes, err := syscall.BytePtrFromString(unitsName)
	if err != nil {
		return fmt.Errorf("failed to convert units name to bytes: %v", err)
	}

	// Call SimConnect_AddToDataDefinition
	// HRESULT SimConnect_AddToDataDefinition(HANDLE hSimConnect, SIMCONNECT_DATA_DEFINITION_ID DefineID,
	//                                        const char* DatumName, const char* UnitsName,
	//                                        SIMCONNECT_DATATYPE DatumType, float fEpsilon, DWORD DatumID)
	r1, _, _ := proc.Call(
		c.handle,                                // hSimConnect
		uintptr(defineID),                       // DefineID
		uintptr(unsafe.Pointer(datumNameBytes)), // DatumName
		uintptr(unsafe.Pointer(unitsNameBytes)), // UnitsName
		uintptr(datumType),                      // DatumType
		uintptr(0),                              // fEpsilon (0.0 for exact match)
		uintptr(0),                              // DatumID (0 for automatic assignment)
	)

	hresult := uint32(r1)
	if !IsHRESULTSuccess(hresult) {
		return NewSimConnectError("SimConnect_AddToDataDefinition", hresult, GetHRESULTMessage(hresult))
	}

	return nil
}

// RequestDataOnSimObject requests data for the specified simulation object
// Implements SimConnect_RequestDataOnSimObject function
func (c *Client) RequestDataOnSimObject(requestID SimObjectDataRequestID, defineID DataDefinitionID, objectID SIMCONNECT_OBJECT_ID, period SIMCONNECT_PERIOD) error {
	if !c.isOpen {
		return fmt.Errorf("client is not open")
	}

	// Get the SimConnect_RequestDataOnSimObject function from DLL
	proc := c.dll.NewProc("SimConnect_RequestDataOnSimObject")

	// Call SimConnect_RequestDataOnSimObject
	// HRESULT SimConnect_RequestDataOnSimObject(HANDLE hSimConnect, SIMCONNECT_DATA_REQUEST_ID RequestID,
	//                                           SIMCONNECT_DATA_DEFINITION_ID DefineID, SIMCONNECT_OBJECT_ID ObjectID,
	//                                           SIMCONNECT_PERIOD Period, SIMCONNECT_DATA_REQUEST_FLAG Flags,
	//                                           DWORD origin, DWORD interval, DWORD limit)
	r1, _, _ := proc.Call(
		c.handle,           // hSimConnect
		uintptr(requestID), // RequestID
		uintptr(defineID),  // DefineID
		uintptr(objectID),  // ObjectID
		uintptr(period),    // Period
		uintptr(0),         // Flags (0 for default)
		uintptr(0),         // origin (0 for default)
		uintptr(0),         // interval (0 for default)
		uintptr(0),         // limit (0 for default)
	)

	hresult := uint32(r1)
	if !IsHRESULTSuccess(hresult) {
		return NewSimConnectError("SimConnect_RequestDataOnSimObject", hresult, GetHRESULTMessage(hresult))
	}

	return nil
}

// GetRawDispatch retrieves the next message from SimConnect as raw bytes
// Implements SimConnect_GetNextDispatch function returning raw data
func (c *Client) GetRawDispatch() ([]byte, error) {
	if !c.isOpen {
		return nil, fmt.Errorf("client is not open")
	}

	// Get the SimConnect_GetNextDispatch function from DLL
	proc := c.dll.NewProc("SimConnect_GetNextDispatch")

	var pData uintptr
	var cbData uint32

	// Call SimConnect_GetNextDispatch
	// HRESULT SimConnect_GetNextDispatch(HANDLE hSimConnect, SIMCONNECT_RECV** ppData, DWORD* pcbData)
	r1, _, _ := proc.Call(
		c.handle,                         // hSimConnect
		uintptr(unsafe.Pointer(&pData)),  // ppData
		uintptr(unsafe.Pointer(&cbData)), // pcbData
	)

	hresult := uint32(r1)

	// Handle E_FAIL as "no data available" - this is normal behavior
	if hresult == E_FAIL {
		return nil, nil // No message available in queue
	}

	if !IsHRESULTSuccess(hresult) {
		return nil, NewSimConnectError("SimConnect_GetNextDispatch", hresult, GetHRESULTMessage(hresult))
	}

	// Check if we have data
	if pData == 0 || cbData == 0 {
		return nil, nil // No message available
	}

	// Copy the data from the SimConnect-managed memory to our own buffer
	buffer := make([]byte, cbData)
	for i := uint32(0); i < cbData; i++ {
		buffer[i] = *(*byte)(unsafe.Pointer(pData + uintptr(i)))
	}

	return buffer, nil
}
