# SimConnect Client API

The `Client` is the core component for establishing and managing connections to Microsoft Flight Simulator 2024 via SimConnect.

## Overview

```go
type Client struct {
    // Private fields for connection management
}
```

The Client handles:
- SimConnect.dll loading and initialization
- Connection establishment and management
- Low-level SimConnect API calls
- Resource cleanup

## Constructor Functions

### NewClient

```go
func NewClient(applicationName string) *Client
```

Creates a new SimConnect client with automatic DLL path detection.

**Parameters:**
- `applicationName` - Name of your application (appears in MSFS developer mode)

**Example:**
```go
client := client.NewClient("MyFlightApp")
```

### NewClientWithDLLPath

```go
func NewClientWithDLLPath(applicationName, dllPath string) *Client
```

Creates a new SimConnect client with custom DLL path.

**Parameters:**
- `applicationName` - Name of your application
- `dllPath` - Full path to SimConnect.dll

**Example:**
```go
client := client.NewClientWithDLLPath("MyApp", "C:\\Custom\\SimConnect.dll")
```

## Connection Management

### Open

```go
func (c *Client) Open() error
```

Establishes connection to SimConnect. Must be called before any other operations.

**Returns:**
- `error` - Connection error, or nil if successful

**Example:**
```go
if err := client.Open(); err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

### Close

```go
func (c *Client) Close() error
```

Closes the SimConnect connection and releases resources. Should be deferred after successful Open().

**Returns:**
- `error` - Error during cleanup, or nil if successful

**Example:**
```go
defer client.Close()
```

### IsOpen

```go
func (c *Client) IsOpen() bool
```

Checks if the client is currently connected to SimConnect.

**Returns:**
- `bool` - true if connected, false otherwise

## Low-Level SimConnect Operations

### AddToDataDefinition

```go
func (c *Client) AddToDataDefinition(defineID DataDefinitionID, datumName, unitsName string, datumType DatumType) error
```

Adds a variable to a SimConnect data definition.

**Parameters:**
- `defineID` - Unique identifier for the data definition
- `datumName` - SimConnect variable name (e.g., "Plane Altitude")
- `unitsName` - Units for the variable (e.g., "feet")
- `datumType` - Data type (typically SIMCONNECT_DATATYPE_FLOAT64)

### RequestDataOnSimObjectWithFlags

```go
func (c *Client) RequestDataOnSimObjectWithFlags(requestID SimObjectDataRequestID, defineID DataDefinitionID, objectID ObjectID, period Period, flags DataRequestFlag, origin, interval, limit uint32) error
```

Requests data from SimConnect with specific flags and parameters.

### SetFloat64OnSimObject

```go
func (c *Client) SetFloat64OnSimObject(defineID DataDefinitionID, objectID ObjectID, value float64) error
```

Sets a float64 value on a simulation object (SetData functionality).

**Parameters:**
- `defineID` - Data definition ID for the variable
- `objectID` - Object ID (typically SIMCONNECT_OBJECT_ID_USER)
- `value` - Value to set

### GetRawDispatch

```go
func (c *Client) GetRawDispatch() ([]byte, error)
```

Retrieves raw dispatch data from SimConnect for processing.

## Error Handling

The Client may return these error types:

- **Connection Errors**: DLL loading, SimConnect initialization failures
- **API Errors**: Invalid parameters, SimConnect API call failures  
- **State Errors**: Operations called when not connected

Always check error returns and implement appropriate error handling:

```go
if err := client.Open(); err != nil {
    if strings.Contains(err.Error(), "DLL") {
        log.Fatal("SimConnect.dll not found - check MSFS installation")
    }
    log.Fatalf("Connection failed: %v", err)
}
```

## Usage Patterns

### Basic Connection Pattern

```go
func connectToSimulator() *client.Client {
    simClient := client.NewClient("MyApp")
    
    if err := simClient.Open(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    return simClient
}

func main() {
    simClient := connectToSimulator()
    defer simClient.Close()
    
    // Use client for operations...
}
```

### Connection with Retry

```go
func connectWithRetry(maxRetries int) (*client.Client, error) {
    simClient := client.NewClient("MyApp")
    
    for i := 0; i < maxRetries; i++ {
        if err := simClient.Open(); err == nil {
            return simClient, nil
        }
        
        log.Printf("Connection attempt %d failed, retrying...", i+1)
        time.Sleep(time.Second * 2)
    }
    
    return nil, fmt.Errorf("failed to connect after %d attempts", maxRetries)
}
```

## Thread Safety

The Client is **not thread-safe**. If you need to use it from multiple goroutines, implement your own synchronization. However, the recommended pattern is to use a single Client instance with the FlightDataManager, which provides thread-safe operations.

## Best Practices

1. **Always defer Close()** after successful Open()
2. **Check IsOpen()** before operations if connection state is uncertain
3. **Handle connection errors gracefully** - MSFS may not be running
4. **Use FlightDataManager** for high-level operations instead of direct Client usage
5. **Initialize once** - create one Client instance per application

## See Also

- [Flight Data Manager API](flight-data-manager.md) - High-level data management
- [Error Handling](errors.md) - Comprehensive error handling strategies
- [Getting Started](../getting-started.md) - Basic usage examples
