# Real-Time Flight Monitor Demo

**Location:** `examples/flight_monitor_demo/`

This comprehensive demo showcases advanced system events functionality with bidirectional validation, interactive testing, and real-time state monitoring of Microsoft Flight Simulator.

## Overview

The Flight Monitor Demo is a production-quality example that demonstrates:

- **Real-time Event Monitoring** - Live display of all system events with timestamps
- **Bidirectional Validation** - Both receiving events AND sending commands to trigger changes
- **Interactive Control Interface** - Command-line interface for testing various scenarios
- **State Tracking** - Monitor simulation state changes in real-time
- **Integration Testing** - SystemEventManager + FlightDataManager running concurrently
- **Comprehensive Validation** - Automated tests for system integrity

## Features

### 🎮 Real-Time Monitoring
- Live event stream with timestamps and detailed information
- Visual indicators for different event types (emojis and formatting)
- Periodic status updates every 10 seconds
- Event frequency analysis and statistics

### 🔄 Bidirectional Testing
- **Receive Events** - Monitor all system events from the simulator
- **Send Commands** - Trigger state changes and validate responses
- **State Validation** - Compare expected vs actual state changes
- **Integration Verification** - Test concurrent manager operation

### 📊 Interactive Dashboard
- Real-time flight data display
- Event statistics and breakdowns
- Simulation state tracking
- Performance monitoring (FPS, event rates)

### 🛠️ Testing Commands
Interactive commands for comprehensive validation:

| Command | Description | Example |
|---------|-------------|---------|
| `status` | Show detailed statistics and state | `status` |
| `data` | Display current flight data | `data` |
| `camera X` | Change camera view (test writable vars) | `camera 4` |
| `throttle X` | Set throttle percentage (test SetData) | `throttle 75` |
| `test` | Run automated validation tests | `test` |
| `help` | Show available commands | `help` |
| `quit` | Exit the monitor | `quit` |

## Quick Start

### Prerequisites
- Microsoft Flight Simulator 2024 running
- Aircraft loaded (any aircraft)
- SimConnect enabled in MSFS settings

### Build and Run
```powershell
# Build the demo
go build -o ./bin/flight-monitor-demo.exe ./examples/flight_monitor_demo/

# Run the demo
./bin/flight-monitor-demo.exe
```

### First Run
1. **Start MSFS 2024** and load any aircraft
2. **Run the demo** - it will connect automatically
3. **Watch events** appear in real-time
4. **Try interactive commands** to test bidirectional functionality
5. **Test manual actions** in MSFS (pause, camera changes, etc.)

## Testing Scenarios

### Automated Tests
Run `test` command to validate:
- ✅ Event subscription status
- ✅ Manager integration
- ✅ State consistency
- ✅ Data variable access
- ✅ Writable variables functionality

### Manual Validation Tests

#### 1. Pause Event Testing
```
Action: Press ESC or PAUSE in MSFS
Expected: See Pause/Paused/Unpaused events
Validation: status command shows correct pause state
```

#### 2. Camera Control Testing
```
Action: Type 'camera 4' in demo
Expected: Camera switches to external view
Validation: ViewChanged event fires, aircraft view changes
```

#### 3. Throttle Control Testing
```
Action: Type 'throttle 75' in demo
Expected: Throttle moves to 75%
Validation: Flight data shows updated throttle position
```

#### 4. Flight Loading Testing
```
Action: Load different flight in MSFS
Expected: FlightLoaded event with filename
Validation: status shows new flight name
```

#### 5. Aircraft Change Testing
```
Action: Change aircraft in MSFS
Expected: AircraftLoaded event with filename
Validation: status shows new aircraft name
```

## Event Types Monitored

### Timer Events
- **Timer1Sec** ⏱️ - Every second (regular monitoring)
- **Timer4Sec** ⏰ - Every 4 seconds (periodic checks)
- **Timer6Hz** ⚡ - 6 times per second (high frequency)
- **Frame** 🎬 - Every visual frame (performance monitoring)

### Simulation State
- **SimStart** ▶️ - Simulation started
- **SimStop** ⏹️ - Simulation stopped
- **Sim** 🎮 - Simulation running state

### Pause Events
- **Pause** ⏸️ - Pause state changed
- **Paused** 🔸 - Currently paused
- **Unpaused** ▶️ - Currently running
- **PauseEx** ⏸️ - Extended pause info with flags

### Flight Events
- **FlightLoaded** ✈️ - New flight loaded
- **FlightSaved** 💾 - Flight saved
- **AircraftLoaded** 🛩️ - Aircraft changed

### System Events
- **PositionChanged** 📍 - Aircraft position changed
- **ViewChanged** 👁️ - Camera view changed
- **Sound** 🔊 - Master sound toggle
- **Crashed** 💥 - Aircraft crashed
- **CrashReset** 🔄 - Crash state reset

## Sample Output

### Startup Sequence
```
=== REAL-TIME FLIGHT MONITOR DEMO ===
STEP 1: Connecting to Microsoft Flight Simulator...
✅ Successfully connected to SimConnect!
STEP 2: Setting up event and data managers...
  ✅ Subscribed: Timer1Sec
  ✅ Subscribed: Pause
  ✅ Subscribed: ViewChanged
✅ Managers configured successfully!
STEP 3: Starting real-time monitoring...
✅ Real-time monitoring active!
```

### Live Event Stream
```
[15:04:05.123] ⏱️ Timer1Sec: Data=1
[15:04:06.234] 📷 ViewChanged: Virtual Cockpit (Data=0x2)
[15:04:07.345] ⏸️ Pause: ON (Data=1)
[15:04:08.456] 🔸 Paused: Data=1
[15:04:10.567] ▶️ Unpaused: Data=0
```

### Status Display
```
=== DETAILED FLIGHT MONITOR STATUS ===
⏱️ Monitoring Duration: 45.3 seconds
📡 Total Events Received: 127
📊 Events Per Second: 2.80

📈 Event Breakdown:
   Timer1Sec           :   45 (35.4%)
   Timer4Sec           :   11 (8.7%)
   Frame               :   60 (47.2%)
   Pause               :    3 (2.4%)
   ViewChanged         :    2 (1.6%)
```

## Architecture

### Component Structure
```
FlightMonitor
├── SimConnect Client (connection management)
├── SystemEventManager (event processing)
├── FlightDataManager (variable data)
├── Dashboard (statistics & display)
└── StateTracker (state monitoring)
```

### Thread Safety
- All components use proper mutex locking
- Concurrent event processing and data updates
- Safe command handling during monitoring

### Error Handling
- Graceful degradation when MSFS unavailable
- Comprehensive error monitoring and reporting
- Recovery mechanisms for connection issues

## Integration Validation

The demo validates successful integration by:

1. **Concurrent Operation** - Both managers run simultaneously
2. **Shared Connection** - Single SimConnect client for both managers
3. **No Conflicts** - Event processing doesn't interfere with data updates
4. **Performance** - Maintains high event rates without degradation
5. **Bidirectional Control** - Can both receive and send data

## Troubleshooting

### No Events Received
```
Total Events Received: 0
```
**Solutions:**
- Ensure MSFS is running with aircraft loaded
- Check SimConnect is enabled in MSFS settings
- Verify correct DLL path for MSFS 2024 SDK
- Try manual actions (pause/unpause) to generate events

### Command Failures
```
❌ Failed to set camera state: client is not open
```
**Solutions:**
- Check MSFS connection status
- Ensure aircraft is loaded and active
- Verify writable variables are properly configured

### High CPU Usage
```
Events Per Second: 120.5
```
**Note:** High event rates (especially Frame events) are normal. Use Ctrl+C to exit if needed.

## Advanced Usage

### Custom Event Filtering
Modify `createEventHandler()` to add custom filtering:

```go
if eventName == "Frame" && event.Data < 30 {
    // Only log low frame rates
    fmt.Printf("⚠️ Low FPS: %d\n", event.Data)
}
```

### Extended State Tracking
Add custom state variables to `StateTracker`:

```go
type StateTracker struct {
    // ... existing fields ...
    customState map[string]interface{}
}
```

### Performance Monitoring
Track event processing performance:

```go
start := time.Now()
// Process event
duration := time.Since(start)
if duration > time.Millisecond {
    log.Printf("Slow event processing: %v", duration)
}
```

## Production Considerations

### Resource Management
- Monitor memory usage during long runs
- Implement event rate limiting if needed
- Use buffered channels to prevent blocking

### Error Recovery
- Implement automatic reconnection logic
- Handle SimConnect service restarts
- Graceful degradation when simulator unavailable

### Logging
- Add structured logging for production deployment
- Implement log rotation for long-running sessions
- Monitor error rates and performance metrics

## Related Documentation

- [SystemEventManager API](../../docs/api/system-events.md) - Complete API reference
- [FlightDataManager API](../../docs/api/flight-data-manager.md) - Variable data management
- [Client API](../../docs/api/client.md) - Core SimConnect functionality
- [System Events Example](../../docs/examples/system-events.md) - Basic usage patterns
- [Troubleshooting Guide](../../docs/advanced/troubleshooting.md) - Common issues

## Validation Results

Upon successful completion, you should see:

```
🎉 Flight Monitor Demo completed successfully!
✅ System events implementation validated with live simulator
📊 FINAL SESSION STATISTICS:
   Duration: 120.5 seconds
   Total Events: 340
   Avg Events/sec: 2.82
```

This confirms that the system events implementation is working correctly with bidirectional validation in a real simulator environment.
