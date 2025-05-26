# SimConnect SetDataOnSimObject Implementation

This document describes the implementation of `SetDataOnSimObject` functionality in the go-simconnect library, enabling both reading and writing of simulation variables.

## Overview

The implementation allows you to:
- ✅ **Read** simulation variables (existing functionality)
- ✅ **Write** simulation variables (new functionality)
- ✅ **Control which variables are writable** (safety feature)
- ✅ **Use both individual and batch operations**

## Architecture Decision: Non-Tagged Mode

### The Question: Tagged vs Non-Tagged Data Setting

SimConnect supports two modes for setting data:
- **Non-Tagged Mode (Default)**: Replace entire data definition with new data
- **Tagged Mode**: Send data in tagged format for selective updates

### Our Decision: Non-Tagged Mode

**Why we chose non-tagged mode:**

1. **Individual Variable Architecture**: Our system creates separate `DataDefinitionID` for each variable
2. **Granular Control**: We want to set individual variables independently  
3. **Simplicity**: Non-tagged mode is simpler and more predictable
4. **Consistency**: Matches our existing read architecture

**Example of our approach:**
```go
// Each variable gets its own data definition
defineID := DataDefinitionID(1)  // Throttle
defineID := DataDefinitionID(2)  // Flaps
defineID := DataDefinitionID(3)  // Gear

// Set individual variables using non-tagged mode
fdm.SetVariable("Throttle Position", 75.0)  // Only affects throttle
fdm.SetVariable("Flaps Position", 25.0)     // Only affects flaps
```

This approach ensures that setting one variable never accidentally affects another.

## Implementation Details

### 1. Constants Added

```go
// SimConnect data set flags for SetDataOnSimObject
type SIMCONNECT_DATA_SET_FLAG uint32

const (
    SIMCONNECT_DATA_SET_FLAG_DEFAULT SIMCONNECT_DATA_SET_FLAG = 0  // Non-tagged mode
    SIMCONNECT_DATA_SET_FLAG_TAGGED  SIMCONNECT_DATA_SET_FLAG = 1  // Tagged mode
)
```

### 2. Core Client Functions

**Low-level function:**
```go
func (c *Client) SetDataOnSimObject(
    defineID DataDefinitionID, 
    objectID SIMCONNECT_OBJECT_ID, 
    flags SIMCONNECT_DATA_SET_FLAG, 
    data []byte,
) error
```

**Convenience functions:**
```go
func (c *Client) SetFloat64OnSimObject(defineID DataDefinitionID, objectID SIMCONNECT_OBJECT_ID, value float64) error
func (c *Client) SetFloat32OnSimObject(defineID DataDefinitionID, objectID SIMCONNECT_OBJECT_ID, value float32) error  
func (c *Client) SetInt32OnSimObject(defineID DataDefinitionID, objectID SIMCONNECT_OBJECT_ID, value int32) error
```

### 3. FlightDataManager Extensions

**Enhanced FlightVariable:**
```go
type FlightVariable struct {
    Name     string    // Human-readable name
    SimVar   string    // SimConnect variable name
    Units    string    // Units of measurement
    Value    float64   // Current value
    Updated  time.Time // Last update time
    Writable bool      // Whether this variable can be written to (NEW)
}
```

**New methods:**
```go
// Add variables with write capability
func (fdm *FlightDataManager) AddVariableWithWritable(name, simVar, units string, writable bool) error

// Set variables by name
func (fdm *FlightDataManager) SetVariable(name string, value float64) error

// Set variables by index (more efficient)
func (fdm *FlightDataManager) SetVariableByIndex(index int, value float64) error

// Add common writable variables
func (fdm *FlightDataManager) AddWritableStandardVariables() error
```

## Usage Examples

### Basic Usage

```go
// Create client and flight data manager
simClient := client.NewClient("MyApp")
simClient.Open()
fdm := client.NewFlightDataManager(simClient)

// Add writable variables
fdm.AddVariableWithWritable("Throttle Position", "General Eng Throttle Lever Position:1", "percent", true)
fdm.AddVariableWithWritable("Flaps Position", "Flaps Handle Percent", "percent", true)

// Add read-only variables  
fdm.AddVariable("Altitude", "Plane Altitude", "feet")

// Start data collection
fdm.Start()

// Set values
fdm.SetVariable("Throttle Position", 75.0)  // Set throttle to 75%
fdm.SetVariable("Flaps Position", 25.0)     // Set flaps to 25%

// Reading values still works as before
if variable, found := fdm.GetVariable("Altitude"); found {
    fmt.Printf("Current altitude: %.0f feet\n", variable.Value)
}
```

### Advanced Usage

```go
// Add commonly writable variables all at once
fdm.AddWritableStandardVariables()

// More efficient repeated operations using index
throttleIndex := 0  // Assuming throttle is first writable variable
for i := 0; i <= 100; i += 10 {
    fdm.SetVariableByIndex(throttleIndex, float64(i))
    time.Sleep(100 * time.Millisecond)
}
```

### Error Handling

```go
// The system prevents setting read-only variables
err := fdm.SetVariable("Altitude", 5000.0)  
if err != nil {
    fmt.Printf("Expected error: %v\n", err)  // "variable 'Altitude' is not writable"
}

// Check if variable is writable before setting
if variable, found := fdm.GetVariable("Throttle Position"); found && variable.Writable {
    fdm.SetVariable("Throttle Position", 50.0)
}
```

## Commonly Writable Variables

The implementation includes these commonly writable variables:

| Variable Name | SimConnect Variable | Units | Description |
|--------------|-------------------|-------|-------------|
| Throttle Position | General Eng Throttle Lever Position:1 | percent | Engine throttle |
| Flaps Position | Flaps Handle Percent | percent | Wing flaps |
| Gear Position | Gear Handle Position | bool | Landing gear |
| Mixture Position | General Eng Mixture Lever Position:1 | percent | Fuel mixture |
| Autopilot Master | Autopilot Master | bool | Autopilot on/off |
| Autopilot Altitude Lock | Autopilot Altitude Lock | bool | Altitude hold |
| Autopilot Heading Lock | Autopilot Heading Lock | bool | Heading hold |

## Safety Features

1. **Write Protection**: Variables must be explicitly marked as writable
2. **Validation**: Check variable exists and is writable before setting
3. **Error Handling**: Comprehensive error reporting
4. **Individual Control**: Each variable has its own data definition
5. **Thread Safety**: All operations are thread-safe

## Testing

Run the comprehensive test to validate the implementation:

```bash
cd cmd/comprehensive_test
./comprehensive_test.exe
```

The test will:
- Connect to SimConnect
- Add both readable and writable variables
- Demonstrate setting various control surfaces
- Show error handling for read-only variables
- Monitor values in real-time

## Technical Notes

### Why Non-Tagged Mode Works for Us

1. **Separate Data Definitions**: Each variable has its own `DataDefinitionID`
2. **Single Value Operations**: We typically set one variable at a time
3. **Clear Boundaries**: No risk of overwriting other variables
4. **Simplicity**: Easier to understand and debug

### Performance Considerations

1. **Individual Calls**: Each `SetVariable` call results in one SimConnect API call
2. **Efficient Index Access**: Use `SetVariableByIndex` for repeated operations
3. **Batch Operations**: For multiple changes, consider grouping them in time
4. **Thread Safety**: All operations are protected by mutexes

### Limitations

1. **Single Aircraft**: Currently targets `SIMCONNECT_OBJECT_ID_USER` (user's aircraft)
2. **Float64 Focus**: Optimized for float64 values (most flight sim variables)
3. **No Tagged Mode**: Does not support tagged format (by design choice)
4. **SimConnect Dependency**: Requires Microsoft Flight Simulator running

## Future Enhancements

Potential improvements for future versions:
- Support for multiple aircraft (AI aircraft)
- Batch setting operations
- Tagged mode support (if needed)
- More data type helpers (strings, complex structures)
- Validation of value ranges
- Undo/redo functionality
