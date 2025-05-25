package simconnect

import (
	"fmt"
	"unsafe"
)

// SIMCONNECT_RECV_ID constants
const (
	SIMCONNECT_RECV_ID_NULL                   = 0x00000000
	SIMCONNECT_RECV_ID_EXCEPTION              = 0x00000001
	SIMCONNECT_RECV_ID_OPEN                   = 0x00000002
	SIMCONNECT_RECV_ID_QUIT                   = 0x00000003
	SIMCONNECT_RECV_ID_EVENT                  = 0x00000004
	SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE = 0x00000005
	SIMCONNECT_RECV_ID_EVENT_FILENAME         = 0x00000006
	SIMCONNECT_RECV_ID_EVENT_FRAME            = 0x00000007
	SIMCONNECT_RECV_ID_SIMOBJECT_DATA         = 0x00000008
	SIMCONNECT_RECV_ID_SIMOBJECT_DATA_BYTYPE  = 0x00000009
	SIMCONNECT_RECV_ID_WEATHER_OBSERVATION    = 0x0000000A
	SIMCONNECT_RECV_ID_CLOUD_STATE            = 0x0000000B
	SIMCONNECT_RECV_ID_ASSIGNED_OBJECT_ID     = 0x0000000C
	SIMCONNECT_RECV_ID_RESERVED_KEY           = 0x0000000D
	SIMCONNECT_RECV_ID_CUSTOM_ACTION          = 0x0000000E
	SIMCONNECT_RECV_ID_SYSTEM_STATE           = 0x0000000F // This was wrong!
	SIMCONNECT_RECV_ID_CLIENT_DATA            = 0x00000010
	SIMCONNECT_RECV_ID_EVENT_WEATHER_MODE     = 0x00000011
	SIMCONNECT_RECV_ID_AIRPORT_LIST           = 0x00000012
	SIMCONNECT_RECV_ID_VOR_LIST               = 0x00000013
	SIMCONNECT_RECV_ID_NDB_LIST               = 0x00000014
	SIMCONNECT_RECV_ID_WAYPOINT_LIST          = 0x00000015
)

// MAX_PATH constant from Windows
const MAX_PATH = 260

// SIMCONNECT_RECV base structure
type SIMCONNECT_RECV struct {
	DwSize    uint32 // Total size of the returned structure in bytes
	DwVersion uint32 // Version number of the SimConnect server
	DwID      uint32 // ID of the returned structure
}

// SIMCONNECT_RECV_SYSTEM_STATE structure for system state responses
type SIMCONNECT_RECV_SYSTEM_STATE struct {
	SIMCONNECT_RECV                // Inherited base structure
	DwRequestID     uint32         // Client defined request ID
	DwInteger       uint32         // Integer/boolean value
	FFloat          float32        // Float value
	SzString        [MAX_PATH]byte // Null-terminated string
}

// SIMCONNECT_RECV_SIMOBJECT_DATA structure for simulation object data responses
type SIMCONNECT_RECV_SIMOBJECT_DATA struct {
	SIMCONNECT_RECV         // Inherited base structure
	DwRequestID      uint32 // Client defined request ID
	DwObjectID       uint32 // Simulation object ID
	DwDefineID       uint32 // Data definition ID
	DwFlags          uint32 // Flags (reserved)
	DwentrynumberOut uint32 // Entry number (reserved)
	DwoutofOut       uint32 // Out of (reserved)
	DwDefineCount    uint32 // Number of data definitions
	// Data follows this structure - must be cast to appropriate type
}

// ParseSimObjectData parses a SIMCONNECT_RECV_SIMOBJECT_DATA message from raw bytes
func ParseSimObjectData(data []byte) (*SIMCONNECT_RECV_SIMOBJECT_DATA, []byte, error) {
	if len(data) < int(unsafe.Sizeof(SIMCONNECT_RECV_SIMOBJECT_DATA{})) {
		return nil, nil, fmt.Errorf("data too short for SIMCONNECT_RECV_SIMOBJECT_DATA")
	}

	// Cast the data to the structure
	recv := (*SIMCONNECT_RECV_SIMOBJECT_DATA)(unsafe.Pointer(&data[0]))

	// Calculate where the actual simulation data starts
	headerSize := unsafe.Sizeof(SIMCONNECT_RECV_SIMOBJECT_DATA{})
	if len(data) <= int(headerSize) {
		return recv, nil, nil // No data portion
	}

	// Return the header and the remaining data
	simData := data[headerSize:]
	return recv, simData, nil
}

// ParseMessageType returns the message type from raw SimConnect data
func ParseMessageType(data []byte) (uint32, error) {
	if len(data) < int(unsafe.Sizeof(SIMCONNECT_RECV{})) {
		return 0, fmt.Errorf("data too short for SIMCONNECT_RECV")
	}

	recv := (*SIMCONNECT_RECV)(unsafe.Pointer(&data[0]))
	return recv.DwID, nil
}

// SystemStateResponse represents a processed system state response
type SystemStateResponse struct {
	RequestID    DataRequestID
	StringValue  string
	IntegerValue uint32
	FloatValue   float32
	DataType     string // "string", "integer", "float"
}

// GetNextDispatch retrieves the next SimConnect message
func (c *Client) GetNextDispatch() (*SystemStateResponse, error) {
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

	// Handle E_FAIL as "no data available" - this is normal behavior for GetNextDispatch
	if hresult == E_FAIL {
		return nil, nil // No message available in queue
	}

	// Only treat other non-success codes as actual errors
	if !IsHRESULTSuccess(hresult) {
		return nil, NewSimConnectError("SimConnect_GetNextDispatch", hresult, GetHRESULTMessage(hresult))
	}

	// Check if we have data
	if pData == 0 || cbData == 0 {
		return nil, nil // No message available
	}

	// Read the base SIMCONNECT_RECV structure
	recv := (*SIMCONNECT_RECV)(unsafe.Pointer(pData))

	// Check if this is a system state response
	if recv.DwID == SIMCONNECT_RECV_ID_SYSTEM_STATE {
		// Cast to SIMCONNECT_RECV_SYSTEM_STATE
		systemStateRecv := (*SIMCONNECT_RECV_SYSTEM_STATE)(unsafe.Pointer(pData))

		// Convert the response to our Go structure
		response := &SystemStateResponse{
			RequestID:    DataRequestID(systemStateRecv.DwRequestID),
			IntegerValue: systemStateRecv.DwInteger,
			FloatValue:   systemStateRecv.FFloat,
		}

		// Convert the C string to Go string
		response.StringValue = cStringToGoString(systemStateRecv.SzString[:])

		// Determine the primary data type based on content
		if response.StringValue != "" {
			response.DataType = "string"
		} else if systemStateRecv.FFloat != 0.0 {
			response.DataType = "float"
		} else {
			response.DataType = "integer"
		}

		return response, nil
	}

	// Not a system state response - we got another message type
	// This is normal, just return nil to indicate we should continue polling
	return nil, nil
}

// GetNextDispatchDebug retrieves the next SimConnect message with debug information
func (c *Client) GetNextDispatchDebug() (*SystemStateResponse, error) {
	if !c.isOpen {
		return nil, fmt.Errorf("client is not open")
	}

	// Get the SimConnect_GetNextDispatch function from DLL
	proc := c.dll.NewProc("SimConnect_GetNextDispatch")

	var pData uintptr
	var cbData uint32

	// Call SimConnect_GetNextDispatch
	r1, _, _ := proc.Call(
		c.handle,                         // hSimConnect
		uintptr(unsafe.Pointer(&pData)),  // ppData
		uintptr(unsafe.Pointer(&cbData)), // pcbData
	)
	hresult := uint32(r1)

	// Handle E_FAIL as "no data available"
	if hresult == E_FAIL {
		return nil, nil // No message available in queue
	}

	// Only treat other non-success codes as actual errors
	if !IsHRESULTSuccess(hresult) {
		return nil, NewSimConnectError("SimConnect_GetNextDispatch", hresult, GetHRESULTMessage(hresult))
	}

	// Check if we have data
	if pData == 0 || cbData == 0 {
		return nil, nil // No message available
	}

	// Read the base SIMCONNECT_RECV structure
	recv := (*SIMCONNECT_RECV)(unsafe.Pointer(pData))

	// Debug: Print what message type we received
	fmt.Printf("🔍 DEBUG: Received message type: 0x%08X, size: %d bytes\n", recv.DwID, cbData)

	// Check if this is a system state response
	if recv.DwID == SIMCONNECT_RECV_ID_SYSTEM_STATE {
		fmt.Println("✅ Found SYSTEM_STATE message!")
		// Cast to SIMCONNECT_RECV_SYSTEM_STATE
		systemStateRecv := (*SIMCONNECT_RECV_SYSTEM_STATE)(unsafe.Pointer(pData))

		// Convert the response to our Go structure
		response := &SystemStateResponse{
			RequestID:    DataRequestID(systemStateRecv.DwRequestID),
			IntegerValue: systemStateRecv.DwInteger,
			FloatValue:   systemStateRecv.FFloat,
		}

		// Convert the C string to Go string
		response.StringValue = cStringToGoString(systemStateRecv.SzString[:])

		// Determine the primary data type based on content
		if response.StringValue != "" {
			response.DataType = "string"
		} else if systemStateRecv.FFloat != 0.0 {
			response.DataType = "float"
		} else {
			response.DataType = "integer"
		}

		return response, nil
	} else { // Print what other message types we're getting
		switch recv.DwID {
		case SIMCONNECT_RECV_ID_NULL:
			fmt.Println("📭 Received NULL message")
		case SIMCONNECT_RECV_ID_EXCEPTION:
			fmt.Println("⚠️  Received EXCEPTION message")
		case SIMCONNECT_RECV_ID_OPEN:
			fmt.Println("🔗 Received OPEN confirmation message")
		case SIMCONNECT_RECV_ID_QUIT:
			fmt.Println("👋 Received QUIT message")
		case SIMCONNECT_RECV_ID_EVENT:
			fmt.Println("📡 Received EVENT message")
		case SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
			fmt.Println("📊 Received SIMOBJECT_DATA message")
		case SIMCONNECT_RECV_ID_CLIENT_DATA:
			fmt.Println("💾 Received CLIENT_DATA message")
		default:
			fmt.Printf("❓ Received unknown message type: 0x%08X\n", recv.DwID)
		}
	}

	// Not a system state response
	return nil, nil
}

// cStringToGoString converts a null-terminated C string byte array to Go string
func cStringToGoString(data []byte) string {
	// Find the null terminator
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data) // No null terminator found, return the whole array
}

// GetSimObjectData retrieves the next SimConnect message and returns simulation object data if found
func (c *Client) GetSimObjectData() (*SIMCONNECT_RECV_SIMOBJECT_DATA, []float64, error) {
	if !c.isOpen {
		return nil, nil, fmt.Errorf("client is not open")
	}

	// Get raw message data
	data, err := c.GetRawDispatch()
	if err != nil {
		return nil, nil, err
	}

	if data == nil {
		return nil, nil, nil // No message available
	}

	// Check message type
	msgType, err := ParseMessageType(data)
	if err != nil {
		return nil, nil, err
	}

	// Only process SIMOBJECT_DATA messages
	if msgType != SIMCONNECT_RECV_ID_SIMOBJECT_DATA {
		return nil, nil, nil // Not a simulation object data message
	}

	// Parse the simulation object data
	header, simData, err := ParseSimObjectData(data)
	if err != nil {
		return nil, nil, err
	}

	if simData == nil || len(simData) == 0 {
		return header, nil, nil // Header only, no data
	}

	// Parse the simulation data as an array of float64 values
	// Each float64 is 8 bytes
	const float64Size = 8
	numFloats := len(simData) / float64Size

	if numFloats == 0 {
		return header, nil, nil
	}

	floats := make([]float64, numFloats)
	for i := 0; i < numFloats; i++ {
		offset := i * float64Size
		if offset+float64Size <= len(simData) {
			// Convert bytes to float64
			floats[i] = *(*float64)(unsafe.Pointer(&simData[offset]))
		}
	}

	return header, floats, nil
}
