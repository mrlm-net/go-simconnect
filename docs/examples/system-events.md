# System Events Monitoring Example

**Location:** `examples/system_events_comprehensive/`

This comprehensive example demonstrates the complete system events functionality of go-simconnect, showing how to monitor Microsoft Flight Simulator state changes in real-time without polling.

## Overview

The system events example showcases:
- Event-driven notifications from the simulator
- Real-time monitoring of simulation state changes
- Integration with FlightDataManager for combined functionality
- Event statistics and performance tracking
- Professional error handling and graceful shutdown

## Features

### Event Types Demonstrated
- **Timer Events** - Regular interval notifications (1sec, 4sec, 6Hz)
- **Simulation State** - Start/stop/pause/unpause events
- **Flight Events** - Flight loaded, saved, aircraft changes
- **System Events** - Position changes, view changes, crashes
- **Performance Events** - Frame rate monitoring

### Technical Features
- Thread-safe event processing
- Concurrent operation with FlightDataManager
- Real-time event statistics and monitoring
- Professional logging with timestamps
- Graceful shutdown with Ctrl+C handling
- Error monitoring and reporting

## Quick Start

```bash
# Navigate to the example directory
cd examples/system_events_comprehensive

# Build the example
go build -o ../../bin/system-events-comprehensive.exe .

# Run the example (ensure MSFS is running)
../../bin/system-events-comprehensive.exe
```

## What You'll See

When you run the example, you'll see:

1. **Startup Sequence**
   ```
   === COMPREHENSIVE SYSTEM EVENTS DEMONSTRATION ===
   STEP 1: Creating SimConnect client...
   STEP 2: Connecting to Microsoft Flight Simulator...
   ✅ Successfully connected to SimConnect!
   ```

2. **Event Subscription**
   ```
   STEP 4: Subscribing to system events...
     ✅ Subscribed to: Timer 1sec
     ✅ Subscribed to: Timer 4sec
     ✅ Subscribed to: Sim Start
     ✅ Subscribed to: Pause
     ✅ Subscribed to: Flight Loaded
   ```

3. **Real-time Event Monitoring**
   ```
   === REAL-TIME EVENT MONITORING ===
   [15:04:05] 🔔 Timer 1sec: Data=1
   [15:04:06] 🔔 Timer 1sec: Data=1
   [15:04:08] 🔔 Timer 4sec: Data=1
   [15:04:09] 🔸 Pause: Data=1
   [15:04:12] ▶️ Unpaused: Data=0
   ```

4. **Periodic Statistics**
   ```
   --- EVENT STATISTICS ---
   Monitoring duration: 30.2 seconds
   Total system events received: 47
   Event breakdown:
     Timer 1sec          :   30 (63.8%)
     Timer 4sec          :    7 (14.9%)
     Pause               :    3 (6.4%)
     Flight Loaded       :    1 (2.1%)
   ```

## Code Structure

### Main Components

#### Event Counter
```go
type EventCounter struct {
    mu    sync.RWMutex
    total map[string]int
}
```
Thread-safe event counting for statistics.

#### Event Callbacks
```go
createEventCallback := func(eventName string) client.SystemEventCallback {
    return func(event client.SystemEventData) {
        eventCounter.Increment(eventName)
        timestamp := time.Now().Format("15:04:05")
        
        // Format based on event type
        switch event.EventType {
        case "basic":
            fmt.Printf("[%s] 🔔 %s: Data=%d\n", timestamp, eventName, event.Data)
        case "filename":
            fmt.Printf("[%s] 📄 %s: File=%s, Data=%d\n", 
                timestamp, eventName, event.Filename, event.Data)
        // ... other types
        }
    }
}
```

#### Integration with FlightDataManager
```go
// Create both managers
eventManager := client.NewSystemEventManager(simClient)
fdm := client.NewFlightDataManager(simClient)

// Configure and start both
fdm.AddVariable("Altitude", "INDICATED ALTITUDE", "feet")
fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")

// Both run concurrently
fdm.Start()
eventManager.Start()
```

## Event Subscription Details

### Events Monitored

| Event Name | Description | Frequency | Use Case |
|------------|-------------|-----------|----------|
| Timer 1sec | Every second | High | Regular monitoring |
| Timer 4sec | Every 4 seconds | Medium | Periodic checks |
| Timer 6Hz | 6 times per second | Very High | Performance monitoring |
| Sim Start | Simulation started | Rare | Initialization |
| Sim Stop | Simulation stopped | Rare | Cleanup |
| Pause | Pause state changed | Low | State tracking |
| Paused | Currently paused | Low | UI updates |
| Unpaused | Currently running | Low | Resume operations |
| Flight Loaded | New flight loaded | Rare | Flight tracking |
| Aircraft Loaded | Aircraft changed | Rare | Aircraft tracking |
| Position Changed | Aircraft moved | Medium | Position tracking |
| View Changed | Camera view changed | Low | View tracking |
| Frame | Frame rendered | Very High | Performance analysis |

### Event Types

#### Basic Events
Most common event type with simple data:
```go
func(event client.SystemEventData) {
    fmt.Printf("Event: %d\n", event.Data)
}
```

#### Filename Events  
Events that include file information:
```go
func(event client.SystemEventData) {
    if event.EventType == "filename" {
        fmt.Printf("File loaded: %s\n", event.Filename)
    }
}
```

#### Object Events
Events for simulation objects:
```go
func(event client.SystemEventData) {
    if event.EventType == "object" {
        fmt.Printf("Object %d added/removed\n", event.ObjectID)
    }
}
```

## Testing the Example

### Triggering Events

To see different events in action:

1. **Pause Events** - Press ESC or PAUSE key in MSFS
2. **Flight Events** - Load a different flight from the main menu
3. **Aircraft Events** - Change aircraft from the hangar
4. **Position Events** - Use slew mode to move the aircraft
5. **View Events** - Change camera views (external, cockpit, etc.)

### Expected Event Patterns

- **Timer events** fire continuously while connected
- **Pause/unpause** events occur when you pause/resume the sim
- **Flight loaded** events occur when loading flights or scenarios
- **Position changed** events occur during aircraft movement
- **View changed** events occur when switching camera perspectives

## Integration Validation

The example demonstrates successful integration by:

1. **Concurrent Operation** - Both SystemEventManager and FlightDataManager run simultaneously
2. **Shared Connection** - Both use the same SimConnect client instance
3. **No Conflicts** - Event processing doesn't interfere with flight data updates
4. **Error Handling** - Both managers report errors independently
5. **Performance** - No performance degradation from running both

## Performance Monitoring

### Event Rate Analysis
The example tracks:
- Total events received
- Events per second average
- Event breakdown by type
- Monitoring duration

### Memory Usage
- Event callbacks are lightweight
- No memory leaks from event subscriptions
- Proper cleanup on shutdown

### CPU Impact
- Event processing is efficient
- Background goroutines don't block main thread
- Minimal CPU overhead for event handling

## Troubleshooting

### Common Issues

#### No Events Received
```
Total system events received: 0
```
**Solutions:**
- Ensure MSFS is running and loaded with an aircraft
- Check that SimConnect is enabled in MSFS settings
- Verify the correct DLL path for MSFS 2024 SDK
- Try pausing/unpausing the simulation to generate events

#### High Frequency Events
```
[15:04:05] 🖼️ Frame: FrameRate=60, Data=60
[15:04:05] 🖼️ Frame: FrameRate=60, Data=60
```
**Note:** Frame events fire at display refresh rate (60+ Hz). This is normal behavior.

#### Connection Issues
```
Failed to connect to SimConnect: connection refused
```
**Solutions:**
- Ensure MSFS is running
- Check Windows Firewall settings
- Verify SimConnect SDK installation
- Try running as Administrator

### Event Debugging

Enable detailed logging by modifying the callback:
```go
func(event client.SystemEventData) {
    log.Printf("DEBUG: Event=%s, Type=%s, Data=%d, ObjectID=%d, Filename=%s",
        eventName, event.EventType, event.Data, event.ObjectID, event.Filename)
}
```

## Advanced Usage

### Custom Event Filtering
```go
pauseHandler := func(event client.SystemEventData) {
    switch event.Data {
    case uint32(client.SIMCONNECT_STATE_ON):
        fmt.Println("Simulation PAUSED")
    case uint32(client.SIMCONNECT_STATE_OFF):
        fmt.Println("Simulation RESUMED")
    }
}
```

### Event Rate Limiting
```go
var lastFrameEvent time.Time
frameHandler := func(event client.SystemEventData) {
    now := time.Now()
    if now.Sub(lastFrameEvent) > 100*time.Millisecond {
        fmt.Printf("Frame rate: %d\n", event.Data)
        lastFrameEvent = now
    }
}
```

### State Machine Integration
```go
type SimulatorState struct {
    IsPaused     bool
    FlightLoaded bool
    AircraftName string
}

var state SimulatorState

pauseHandler := func(event client.SystemEventData) {
    state.IsPaused = (event.Data == uint32(client.SIMCONNECT_STATE_ON))
}

flightHandler := func(event client.SystemEventData) {
    if event.EventType == "filename" {
        state.FlightLoaded = true
        // Parse aircraft name from filename if needed
    }
}
```

## Production Considerations

### Error Recovery
```go
go func() {
    for err := range eventManager.GetErrors() {
        if strings.Contains(err.Error(), "connection lost") {
            // Attempt reconnection
            eventManager.Stop()
            // ... reconnection logic
            eventManager.Start()
        }
    }
}()
```

### Resource Management
- Unsubscribe from unused events
- Monitor goroutine count
- Implement graceful shutdown
- Handle network interruptions

### Logging
- Use structured logging in production
- Log event statistics periodically
- Monitor error rates
- Track connection stability

## Next Steps

After running this example:

1. **Study the Code** - Review the implementation patterns
2. **Modify Events** - Try subscribing to different events
3. **Add Features** - Implement custom event handlers
4. **Integration** - Combine with your own applications
5. **Production** - Apply the patterns to real projects

## Related Documentation

- [SystemEventManager API](../../docs/api/system-events.md) - Complete API reference
- [Client API](../../docs/api/client.md) - Core SimConnect functionality
- [FlightDataManager](../../docs/api/flight-data-manager.md) - Variable data management
- [Troubleshooting](../../docs/advanced/troubleshooting.md) - Common issues and solutions
