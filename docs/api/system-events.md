# SystemEventManager API Reference

The SystemEventManager provides event-driven notifications from Microsoft Flight Simulator, allowing applications to respond to simulation state changes without polling.

## Overview

System events enable real-time monitoring of simulation state changes including:
- Timer events (1sec, 4sec, 6Hz intervals)
- Simulation lifecycle (start, stop, pause, unpause)
- Flight events (loaded, saved, aircraft changes)
- System state (crashes, position changes, view changes)
- Performance monitoring (frame rate events)

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() {
    // Create SimConnect client with MSFS 2024 SDK
    dllPath := `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
    simClient := client.NewClientWithDLLPath("EventDemo", dllPath)
    
    if err := simClient.Open(); err != nil {
        log.Fatal(err)
    }
    defer simClient.Close()
    
    // Create SystemEventManager
    eventManager := client.NewSystemEventManager(simClient)
    
    // Subscribe to pause events
    eventID, err := eventManager.SubscribeToEvent(
        client.SystemEventPause,
        func(event client.SystemEventData) {
            fmt.Printf("Simulation paused: %d\n", event.Data)
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start event monitoring
    if err := eventManager.Start(); err != nil {
        log.Fatal(err)
    }
    defer eventManager.Stop()
    
    // Monitor for 30 seconds
    time.Sleep(30 * time.Second)
}
```

## SystemEventManager

### Constructor

#### `NewSystemEventManager(client *Client) *SystemEventManager`

Creates a new SystemEventManager instance.

**Parameters:**
- `client`: Active SimConnect client instance

**Returns:**
- `*SystemEventManager`: New event manager instance

---

### Event Subscription

#### `SubscribeToEvent(eventName string, callback SystemEventCallback) (SIMCONNECT_CLIENT_EVENT_ID, error)`

Subscribe to a specific system event with a callback function.

**Parameters:**
- `eventName`: System event name constant (e.g., `client.SystemEventPause`)
- `callback`: Function to call when event occurs

**Returns:**
- `SIMCONNECT_CLIENT_EVENT_ID`: Unique event ID for this subscription
- `error`: Error if subscription fails

**Example:**
```go
eventID, err := eventManager.SubscribeToEvent(
    client.SystemEventFlightLoaded,
    func(event client.SystemEventData) {
        fmt.Printf("Flight loaded: %s\n", event.Filename)
    },
)
```

#### `SubscribeToCommonEvents(callback SystemEventCallback) error`

Subscribe to commonly used events in a single call.

**Events included:**
- Timer events (1sec, 4sec)
- Pause/unpause events
- Flight loaded/saved events
- Simulation start/stop events

**Parameters:**
- `callback`: Function to call for all events

**Returns:**
- `error`: Error if any subscription fails

#### `UnsubscribeFromEvent(eventID SIMCONNECT_CLIENT_EVENT_ID) error`

Unsubscribe from a specific event.

**Parameters:**
- `eventID`: Event ID returned from `SubscribeToEvent`

**Returns:**
- `error`: Error if unsubscription fails

#### `UnsubscribeAll() error`

Unsubscribe from all events.

**Returns:**
- `error`: Error if any unsubscription fails

---

### Event Manager Control

#### `Start() error`

Start the event monitoring background process.

**Returns:**
- `error`: Error if start fails

#### `Stop()`

Stop the event monitoring and unsubscribe from all events.

#### `IsRunning() bool`

Check if the event manager is currently running.

**Returns:**
- `bool`: True if running, false otherwise

---

### Status and Monitoring

#### `GetSubscribedEvents() map[SIMCONNECT_CLIENT_EVENT_ID]string`

Get currently subscribed events.

**Returns:**
- `map[SIMCONNECT_CLIENT_EVENT_ID]string`: Map of event IDs to event names

#### `GetErrors() <-chan error`

Get error channel for monitoring runtime errors.

**Returns:**
- `<-chan error`: Read-only error channel

**Example:**
```go
go func() {
    for err := range eventManager.GetErrors() {
        log.Printf("Event error: %v", err)
    }
}()
```

---

## System Event Constants

### Timer Events
- `SystemEvent1Sec` - Fires every second
- `SystemEvent4Sec` - Fires every 4 seconds  
- `SystemEvent6Hz` - Fires 6 times per second

### Simulation Events
- `SystemEventSimStart` - Simulation started
- `SystemEventSimStop` - Simulation stopped
- `SystemEventSimPause` - Simulation paused (legacy)
- `SystemEventSimUnpause` - Simulation unpaused (legacy)

### Pause Events
- `SystemEventPause` - Any pause state change
- `SystemEventPaused` - Simulation is paused
- `SystemEventUnpaused` - Simulation is unpaused
- `SystemEventPauseEx` - Extended pause information

### Flight Events
- `SystemEventFlightLoaded` - Flight plan loaded
- `SystemEventFlightSaved` - Flight plan saved
- `SystemEventAircraftLoaded` - Aircraft changed
- `SystemEventPositionChanged` - Aircraft position changed

### System Events
- `SystemEventCrashed` - Aircraft crashed
- `SystemEventCrashReset` - Crash state reset
- `SystemEventView` - Camera view changed
- `SystemEventSound` - Sound state changed

### Advanced Events
- `SystemEventFrame` - Frame rendering (high frequency)
- `SystemEventObjectAdded` - Object added to simulation
- `SystemEventObjectRemoved` - Object removed from simulation

---

## Event Data Types

### `SystemEventData`

Structure containing event information:

```go
type SystemEventData struct {
    EventType string // "basic", "filename", "object", "frame"
    Data      uint32 // Event-specific data
    Filename  string // For filename events
    ObjectID  uint32 // For object events
}
```

### `SystemEventCallback`

Function signature for event callbacks:

```go
type SystemEventCallback func(event SystemEventData)
```

---

## Event Types

### Basic Events
Most events provide a simple data value:
```go
func(event client.SystemEventData) {
    fmt.Printf("Event data: %d\n", event.Data)
}
```

### Filename Events
Events that include a filename (flight loaded, saved):
```go
func(event client.SystemEventData) {
    if event.EventType == "filename" {
        fmt.Printf("File: %s, Data: %d\n", event.Filename, event.Data)
    }
}
```

### Object Events
Events for objects added/removed:
```go
func(event client.SystemEventData) {
    if event.EventType == "object" {
        fmt.Printf("Object ID: %d, Data: %d\n", event.ObjectID, event.Data)
    }
}
```

### Frame Events
High-frequency rendering events:
```go
func(event client.SystemEventData) {
    if event.EventType == "frame" {
        fmt.Printf("Frame rate info: %d\n", event.Data)
    }
}
```

---

## Integration with FlightDataManager

SystemEventManager and FlightDataManager can run concurrently, sharing the same SimConnect connection:

```go
// Create both managers
eventManager := client.NewSystemEventManager(simClient)
flightManager := client.NewFlightDataManager(simClient)

// Configure flight data
flightManager.AddVariable("Altitude", "INDICATED ALTITUDE", "feet")
flightManager.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")

// Configure event monitoring
eventManager.SubscribeToEvent(client.SystemEventPaused, func(event client.SystemEventData) {
    fmt.Println("Simulation paused - flight data updates will pause")
})

// Start both
if err := flightManager.Start(); err != nil {
    log.Fatal(err)
}
if err := eventManager.Start(); err != nil {
    log.Fatal(err)
}

// Both run concurrently without conflicts
defer flightManager.Stop()
defer eventManager.Stop()
```

---

## Error Handling

### Common Errors

1. **Client not open**: Ensure SimConnect client is connected before creating manager
2. **Event already subscribed**: Each event can only have one active subscription
3. **Invalid event name**: Use provided constants for event names
4. **Manager not started**: Call `Start()` before events will be received

### Error Monitoring

```go
// Monitor for runtime errors
go func() {
    for err := range eventManager.GetErrors() {
        switch {
        case strings.Contains(err.Error(), "connection lost"):
            log.Println("SimConnect connection lost, attempting reconnect...")
            // Handle reconnection logic
        case strings.Contains(err.Error(), "subscription failed"):
            log.Printf("Event subscription error: %v", err)
        default:
            log.Printf("Event manager error: %v", err)
        }
    }
}()
```

---

## Performance Considerations

### Event Frequency
- Timer events (especially 6Hz) generate high frequency callbacks
- Frame events can fire at display refresh rate (60+ Hz)
- Consider event filtering or throttling for high-frequency events

### Callback Performance
- Keep callback functions lightweight and fast
- Avoid blocking operations in callbacks
- Use goroutines for heavy processing

### Memory Management
- Event manager automatically handles cleanup on Stop()
- Unsubscribe from unused events to free resources
- Monitor error channel to prevent goroutine leaks

---

## Examples

### Complete Event Monitoring
```go
// Run the comprehensive example
cd examples/system_events_comprehensive
go run main.go
```

### Custom Event Handler
```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() {
    dllPath := `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
    simClient := client.NewClientWithDLLPath("CustomEvents", dllPath)
    
    if err := simClient.Open(); err != nil {
        log.Fatal(err)
    }
    defer simClient.Close()
    
    eventManager := client.NewSystemEventManager(simClient)
    
    // Custom event handler with state tracking
    var isPaused bool
    var flightLoaded bool
    
    pauseHandler := func(event client.SystemEventData) {
        switch event.Data {
        case uint32(client.SIMCONNECT_STATE_ON):
            isPaused = true
            fmt.Println("🔸 Simulation PAUSED")
        case uint32(client.SIMCONNECT_STATE_OFF):
            isPaused = false
            fmt.Println("▶️ Simulation RESUMED")
        }
    }
    
    flightHandler := func(event client.SystemEventData) {
        if event.EventType == "filename" {
            flightLoaded = true
            fmt.Printf("✈️ Flight loaded: %s\n", event.Filename)
        }
    }
    
    // Subscribe to events
    eventManager.SubscribeToEvent(client.SystemEventPaused, pauseHandler)
    eventManager.SubscribeToEvent(client.SystemEventUnpaused, pauseHandler)
    eventManager.SubscribeToEvent(client.SystemEventFlightLoaded, flightHandler)
    
    if err := eventManager.Start(); err != nil {
        log.Fatal(err)
    }
    defer eventManager.Stop()
    
    // Status monitoring loop
    for i := 0; i < 30; i++ {
        fmt.Printf("Status: Paused=%t, FlightLoaded=%t\n", isPaused, flightLoaded)
        time.Sleep(2 * time.Second)
    }
}
```

---

## See Also

- [Client API](client.md) - Core SimConnect functionality
- [FlightDataManager](flight-data-manager.md) - Variable data management
- [System Events Example](../../examples/system_events_comprehensive/) - Complete implementation
- [Troubleshooting](../advanced/troubleshooting.md) - Common issues and solutions
