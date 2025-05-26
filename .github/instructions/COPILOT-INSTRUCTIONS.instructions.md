---
applyTo: '**'
---
Coding standards, domain knowledge, and preferences that AI should follow.

# GitHub Copilot Instructions for go-simconnect

## Project Overview

This is a Go library for Microsoft Flight Simulator SimConnect integration. It provides thread-safe, production-ready access to flight simulation data with both read and write capabilities.

## Critical Project Patterns

### 1. Thread Safety Requirements
- **ALWAYS** use `sync.RWMutex` for shared data structures
- **NEVER** access FlightDataManager fields without proper locking
- Use `RLock()` for read operations and `Lock()` for write operations
- Example pattern:
```go
func (fdm *FlightDataManager) GetVariable(name string) (FlightVariable, bool) {
    fdm.mutex.RLock()
    defer fdm.mutex.RUnlock()
    // ... safe read operation
}
```

### 2. SimConnect Variable Naming
- **ALWAYS** use exact SimConnect variable names (case-sensitive)
- **NEVER** abbreviate or modify SimConnect variable names
- **ALWAYS** use exact unit strings from SimConnect documentation
- Examples:
  - ✅ `"AIRSPEED INDICATED"` with `"knots"`
  - ✅ `"PLANE ALTITUDE"` with `"feet"`
  - ❌ `"SPEED"` or `"ALTITUDE"`

### 3. Writable vs Read-Only Variables
- **ALWAYS** use `AddVariableWithWritable()` for variables that need to be written
- **NEVER** assume all variables are writable - check documentation
- **ALWAYS** validate writable flag before calling `SetVariable()`
- Pattern:
```go
// Add writable variable
fdm.AddVariableWithWritable("Throttle", "General Eng Throttle Lever Position:1", "percent", true)

// Check before setting
if variable, found := fdm.GetVariable("Throttle"); found && variable.Writable {
    fdm.SetVariable("Throttle", 75.0)
}
```

### 4. Data Definition Architecture
- **ALWAYS** create separate data definitions for each variable
- **NEVER** combine multiple variables into single data definition
- Each variable gets unique `DataDefinitionID` and `SimObjectDataRequestID`
- This approach prevents SimConnect exceptions and ensures reliability

### 5. Error Handling Patterns
- **ALWAYS** check for errors when adding variables
- **ALWAYS** implement error channel monitoring for runtime errors
- **ALWAYS** handle connection failures gracefully
- Pattern:
```go
if err := fdm.AddVariable("name", "simVar", "units"); err != nil {
    log.Printf("Warning: Could not add variable: %v", err)
}

// Monitor errors in goroutine
go func() {
    for err := range fdm.GetErrors() {
        log.Printf("Data error: %v", err)
    }
}()
```

### 6. Client Lifecycle Management
- **ALWAYS** defer `client.Close()` immediately after successful `client.Open()`
- **ALWAYS** check connection state before operations
- **NEVER** assume SimConnect is always available
- Pattern:
```go
simClient := client.NewClient("AppName")
if err := simClient.Open(); err != nil {
    return fmt.Errorf("connection failed: %v", err)
}
defer simClient.Close()
```

### 7. FlightDataManager Usage
- **ALWAYS** add variables before calling `Start()`
- **NEVER** add variables while data manager is running
- **ALWAYS** call `Stop()` before adding new variables during runtime
- Pattern:
```go
fdm := client.NewFlightDataManager(simClient)
// Add ALL variables first
fdm.AddVariable("var1", "SIM_VAR_1", "units")
fdm.AddVariable("var2", "SIM_VAR_2", "units")
// Then start
fdm.Start()
```

## Common Mistakes to Avoid

### ❌ Don't Do This
```go
// Starting before adding variables
fdm.Start()
fdm.AddVariable("name", "var", "units") // Will fail

// Not checking writable flag
fdm.SetVariable("Altitude", 5000) // Read-only variable

// Missing thread safety
variables := fdm.variables // Direct access without lock

// Generic variable names
fdm.AddVariable("Speed", "SPEED", "mph") // Wrong var name & units

// Not handling errors
fdm.AddVariable("name", "var", "units") // No error check
```

### ✅ Do This Instead
```go
// Add variables then start
fdm.AddVariable("name", "var", "units")
if err := fdm.Start(); err != nil {
    return err
}

// Check writable before setting
fdm.AddVariableWithWritable("Throttle", "General Eng Throttle Lever Position:1", "percent", true)
if err := fdm.SetVariable("Throttle", 75.0); err != nil {
    log.Printf("Set failed: %v", err)
}

// Use proper API for safe access
variables := fdm.GetAllVariables() // Thread-safe copy

// Use exact SimConnect names
fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")

// Always handle errors
if err := fdm.AddVariable("name", "var", "units"); err != nil {
    log.Printf("Warning: %v", err)
}
```

## File Structure Rules

### Project Layout
- Use `examples/` directory for example code (NOT `cmd/`)
- Use `pkg/client/` for core library code
- Use `docs/` for documentation with modular structure
- **ALWAYS** output builds to `./bin/` directory
- Follow Go project layout standards

### Build Output Organization
- **ALWAYS** build executables to `./bin/` directory
- Use descriptive names for executables in bin folder
- **ALWAYS** use PowerShell command syntax when providing terminal commands
- Example build commands:
```powershell
# Correct build patterns (PowerShell syntax)
go build -o ./bin/camera-test.exe ./examples/camera_test/
go build -o ./bin/web-dashboard.exe ./examples/web_dashboard/
go build -o ./bin/complete-demo.exe ./examples/final_complete_demo_fixed/

# Wrong - don't build in source directories
go build ./examples/camera_test/  # Creates camera_test.exe in source
```

### Documentation Structure
- API reference goes in `docs/api/`
- Examples documentation in `docs/examples/`
- Advanced topics in `docs/advanced/`
- Getting started guide in `docs/getting-started.md`
- Implementation guides in `docs/` root (e.g., `SetDataImplementation.md`)

### Current Documentation Standards
Based on existing structure, follow these patterns:

#### API Documentation (`docs/api/`)
- `client.md` - Core SimConnect client methods and lifecycle
- `flight-data-manager.md` - High-level data management API
- `variables.md` - Complete SimConnect variables reference with examples
- Each file should have: Purpose, Methods, Examples, Thread Safety notes

#### Implementation Guides (`docs/` root)
- `SetDataImplementation.md` - Detailed implementation guides for specific features
- `getting-started.md` - Installation, setup, first application
- Use descriptive filenames matching the feature being documented

#### Example Documentation (`docs/examples/`)
- Document each example with purpose, usage, and expected output
- Include troubleshooting for common issues
- Provide step-by-step instructions for complex examples

#### Advanced Topics (`docs/advanced/`)
- `troubleshooting.md` - Common issues and solutions
- `performance.md` - Optimization techniques
- `architecture.md` - Design patterns and best practices
- `deployment.md` - Production deployment considerations

## Testing Requirements

### Before Committing
- **ALWAYS** test with live MSFS before committing changes
- **ALWAYS** build all examples to `./bin/` directory
- **ALWAYS** validate SetData functionality with real simulator
- **ALWAYS** check thread safety under concurrent access

### Build Validation Process
- Clean bin directory before building: `rm -rf ./bin/; mkdir ./bin/`
- Build all examples to bin folder
- Test executables from bin directory
- Verify no build artifacts remain in source directories

### Camera Control Testing
- Use values 2-6 for camera states (Wing, Cockpit, External, Tail, Tower)
- Test camera switching with live simulator
- Validate smooth transitions between views

### Documentation Testing
- Verify all links work in documentation
- Check code examples compile and run
- Validate API documentation matches actual methods
- Ensure troubleshooting steps are current

## Performance Considerations

### Data Collection
- Use `SIMCONNECT_PERIOD_SECOND` with `SIMCONNECT_DATA_REQUEST_FLAG_CHANGED`
- Individual data definitions perform better than combined definitions
- Target ~20Hz update rate for real-time applications
- Use buffered error channels (capacity 10) to prevent blocking

### Memory Management
- Return copies of data structures, not direct references
- Use `defer` for proper resource cleanup
- Monitor goroutine lifecycle in long-running applications

## SimConnect Integration Notes

### DLL Path Handling
- Support custom DLL paths for MSFS 2024 SDK
- Default to system-registered SimConnect if no path specified
- Example: `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`

### Connection Patterns
- Implement retry logic for production applications
- Handle "SimConnect not available" gracefully
- Support graceful degradation when MSFS not running

## Code Style Guidelines

### Naming Conventions
- Use descriptive variable names that match SimConnect purpose
- Use consistent error message formats
- Follow Go naming conventions for public/private members

### Documentation
- Include usage examples in all API documentation
- Document thread safety characteristics
- Provide troubleshooting guidance for common issues

## When Making Changes

1. **Read existing code patterns** before implementing new features
2. **Maintain thread safety** in all new code
3. **Test with live simulator** before submitting changes
4. **Update documentation** to reflect changes
5. **Follow existing error handling patterns**
6. **Preserve backward compatibility** when possible

## Example Validation Commands

```powershell
# Build all examples (PowerShell syntax)
go build -o ./bin/camera-test.exe ./examples/camera_test/
go build -o ./bin/web-dashboard.exe ./examples/web_dashboard/
go build -o ./bin/complete-demo.exe ./examples/final_complete_demo_fixed/

# Test with live MSFS
./bin/camera-test.exe

# Validate documentation structure
Get-ChildItem docs/api/, docs/examples/, docs/advanced/
```

Remember: This library is production-ready and used in real applications. Always prioritize reliability, thread safety, and proper error handling over quick implementations.

## Official Documentation References

For authoritative SimConnect information, always reference the official Microsoft documentation:

- **Primary Reference**: [SimConnect SDK Documentation](https://docs.flightsimulator.com/html/Programming_Tools/SimConnect/SimConnect_SDK.htm)
- **Variable Reference**: [Simulation Variables](https://docs.flightsimulator.com/html/Programming_Tools/SimVars/Simulation_Variables.htm)
- **Event Reference**: [Key Events](https://docs.flightsimulator.com/html/Programming_Tools/Event_IDs/Event_IDs.htm)
- **Units Reference**: [SimConnect Units](https://docs.flightsimulator.com/html/Programming_Tools/SimConnect/SimConnect_API_Reference.htm#simconnect_datatype_units)

When in doubt about variable names, units, or writable status, always check the official Microsoft Flight Simulator SDK documentation.
