# FlightDataManager API Reference

The FlightDataManager provides a high-level interface for managing real-time flight simulation data collection and variable monitoring.

## Overview

The FlightDataManager simplifies the process of:
- Adding simulation variables to monitor
- Starting/stopping real-time data collection
- Reading current variable values
- Setting writable variable values
- Managing data collection statistics and errors

## Constructor

### NewFlightDataManager

```go
func NewFlightDataManager(client *Client) *FlightDataManager
```

Creates a new FlightDataManager instance.

**Parameters:**
- `client` - A connected SimConnect client instance

**Returns:**
- `*FlightDataManager` - New FlightDataManager instance

## Variable Management

### AddVariable

```go
func (fdm *FlightDataManager) AddVariable(name, simVar, units string) error
```

Adds a read-only simulation variable to be tracked.

**Parameters:**
- `name` - Human-readable name for the variable
- `simVar` - SimConnect variable name (e.g., "AIRSPEED INDICATED")
- `units` - Units of measurement (e.g., "knots", "feet", "degrees")

**Returns:**
- `error` - Error if variable cannot be added

### AddVariableWithWritable

```go
func (fdm *FlightDataManager) AddVariableWithWritable(name, simVar, units string, writable bool) error
```

Adds a simulation variable with write capability specification.

**Parameters:**
- `name` - Human-readable name for the variable
- `simVar` - SimConnect variable name
- `units` - Units of measurement
- `writable` - Whether this variable can be written to

**Returns:**
- `error` - Error if variable cannot be added

## Data Collection Control

### Start

```go
func (fdm *FlightDataManager) Start() error
```

Begins real-time data collection for all added variables.

**Returns:**
- `error` - Error if data collection cannot be started

**Notes:**
- Must have at least one variable added before starting
- Cannot add variables while running
- Uses optimized 1Hz update rate with change detection

### Stop

```go
func (fdm *FlightDataManager) Stop()
```

Stops real-time data collection.

### IsRunning

```go
func (fdm *FlightDataManager) IsRunning() bool
```

Returns whether the data manager is currently collecting data.

**Returns:**
- `bool` - True if data collection is active

## Data Access

### GetVariable

```go
func (fdm *FlightDataManager) GetVariable(name string) (FlightVariable, bool)
```

Returns the current value of a variable by name.

**Parameters:**
- `name` - Human-readable name of the variable

**Returns:**
- `FlightVariable` - Variable data structure
- `bool` - True if variable was found

### GetAllVariables

```go
func (fdm *FlightDataManager) GetAllVariables() []FlightVariable
```

Returns all current variable values.

**Returns:**
- `[]FlightVariable` - Array of all tracked variables

### SetVariable

```go
func (fdm *FlightDataManager) SetVariable(name string, value float64) error
```

Sets the value of a writable simulation variable by name.

**Parameters:**
- `name` - Human-readable name of the variable
- `value` - New value to set

**Returns:**
- `error` - Error if variable cannot be set

**Notes:**
- Variable must be marked as writable
- Variable must exist in the tracked list

### SetVariableByIndex

```go
func (fdm *FlightDataManager) SetVariableByIndex(index int, value float64) error
```

Sets the value using the variable index (more efficient for repeated operations).

**Parameters:**
- `index` - Index of the variable in the tracked list
- `value` - New value to set

**Returns:**
- `error` - Error if variable cannot be set

## Statistics and Monitoring

### GetStats

```go
func (fdm *FlightDataManager) GetStats() (dataCount int64, errorCount int64, lastUpdate time.Time)
```

Returns data collection statistics.

**Returns:**
- `dataCount` - Total number of data updates received
- `errorCount` - Total number of errors encountered
- `lastUpdate` - Timestamp of last successful data update

### GetErrors

```go
func (fdm *FlightDataManager) GetErrors() <-chan error
```

Returns a channel for receiving errors (non-blocking).

**Returns:**
- `<-chan error` - Read-only channel for receiving errors

**Notes:**
- Channel is buffered with capacity of 10
- Errors are dropped if channel is full

## Data Structures

### FlightVariable

```go
type FlightVariable struct {
    Name     string    // Human-readable name
    SimVar   string    // SimConnect variable name
    Units    string    // Units of measurement
    Value    float64   // Current value
    Updated  time.Time // Last update time
    Writable bool      // Whether this variable can be written to
}
```

Represents a simulation variable with its current state and metadata.

## Example Usage

```go
// Create flight data manager
fdm := client.NewFlightDataManager(client)

// Add variables to track
fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
fdm.AddVariable("Altitude", "INDICATED ALTITUDE", "feet")
fdm.AddVariableWithWritable("Camera State", "CAMERA STATE", "number", true)

// Start data collection
if err := fdm.Start(); err != nil {
    log.Fatal(err)
}

// Read data
if variable, found := fdm.GetVariable("Airspeed"); found {
    fmt.Printf("Current airspeed: %.1f %s\n", variable.Value, variable.Units)
}

// Set writable variable
fdm.SetVariable("Camera State", 5.0)

// Stop when done
fdm.Stop()
```

## Thread Safety

The FlightDataManager is thread-safe and can be accessed from multiple goroutines. All public methods use appropriate locking mechanisms to ensure data consistency.

## Performance Notes

- Data collection runs at 1Hz (once per second) by default
- Only changed values are transmitted to reduce network overhead
- Variable lookup by index is more efficient than lookup by name for repeated operations
- Error channel has limited capacity to prevent memory leaks
