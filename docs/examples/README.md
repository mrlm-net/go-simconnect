# Examples Documentation

This directory contains detailed documentation for all examples included in the go-simconnect library. Each example demonstrates different aspects of SimConnect integration and provides working code you can learn from.

## Available Examples

### [Basic Connection](basic-connection.md)
**Location:** `examples/main.go`

A minimal example showing how to establish a SimConnect connection and perform basic operations. Perfect for getting started with the library.

**Features:**
- Basic connection setup
- Simple variable reading
- Error handling
- Clean shutdown

### [Camera Test](camera-test.md)
**Location:** `examples/camera_test/`

Demonstrates real-time camera control using SetData functionality. Shows how to cycle through different camera views in the simulator.

**Features:**
- SetData operations
- Camera state management
- View cycling automation
- Real-time control validation

### [Web Dashboard](web-dashboard.md)
**Location:** `examples/web_dashboard/`

A complete web-based flight data dashboard that displays real-time aircraft information in a browser interface.

**Features:**
- Real-time flight data display
- Web server with REST API
- Modern HTML/CSS/JavaScript frontend
- Multiple aircraft parameters
- Responsive design

### [System Events Monitoring](system-events.md)
**Location:** `examples/system_events_comprehensive/`

Comprehensive demonstration of event-driven notifications from Microsoft Flight Simulator. Shows how to monitor simulation state changes without polling.

**Features:**
- Event subscription and callbacks
- Real-time event monitoring
- Integration with FlightDataManager
- Event statistics and performance tracking
- Multiple event type handling

### [Complete Demo](complete-demo.md)
**Location:** `examples/final_complete_demo_fixed/`

A comprehensive demonstration showing advanced SimConnect features including both data reading and writing capabilities.

**Features:**
- Comprehensive variable monitoring
- Write operations (autopilot, controls)
- Error handling and recovery
- Statistics and performance monitoring
- Production-ready patterns

### [Production Demo](production-demo.md)
**Location:** `examples/final_production_demo/`

An example showing best practices for production deployment of SimConnect applications.

**Features:**
- Configuration management
- Logging and monitoring
- Error recovery strategies
- Performance optimization
- Deployment considerations

### [Testing Framework](testing-framework.md)
**Location:** `examples/test/`

Examples of how to test SimConnect applications, including mocking and integration testing strategies.

**Features:**
- Unit testing patterns
- Mock SimConnect client
- Integration test setup
- Test data generation
- Continuous integration examples

## Getting Started

1. **Start with Basic Connection** - Learn the fundamentals of connecting to SimConnect
2. **Try Camera Test** - Understand SetData operations with immediate visual feedback
3. **Build Web Dashboard** - See how to create user interfaces for flight data
4. **Study Complete Demo** - Learn advanced patterns and best practices
5. **Review Production Demo** - Understand deployment and production considerations

## Running Examples

All examples are located in the `examples/` directory. To run an example:

```bash
# Navigate to the example directory
cd examples/web_dashboard

# Build and run
go build
./web_dashboard.exe
```

### Prerequisites

- Microsoft Flight Simulator running and loaded with an aircraft
- SimConnect SDK properly installed
- Go 1.19 or later

### Common Issues

- **"SimConnect not found"** - Ensure MSFS is running and SimConnect is enabled
- **"Connection refused"** - Check that MSFS allows external connections
- **"Variable not found"** - Verify the aircraft supports the requested variables

## Example Patterns

### Basic Setup Pattern

```go
// Create client
client, err := simconnect.NewClient("MyApp")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Connect
if err := client.Connect(); err != nil {
    log.Fatal(err)
}
```

### FlightDataManager Pattern

```go
// Create flight data manager
fdm := client.NewFlightDataManager(client)

// Add variables
fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")

// Start data collection
if err := fdm.Start(); err != nil {
    log.Fatal(err)
}
defer fdm.Stop()
```

### Error Handling Pattern

```go
// Handle errors from data collection
go func() {
    for err := range fdm.GetErrors() {
        log.Printf("FDM Error: %v", err)
    }
}()
```

## Contributing Examples

If you've created a useful example that demonstrates go-simconnect capabilities:

1. Create a new directory under `examples/`
2. Include a `README.md` with usage instructions
3. Add comprehensive comments to your code
4. Test with multiple aircraft types
5. Submit a pull request

### Example Guidelines

- **Clear Purpose** - Each example should demonstrate specific features
- **Complete Code** - Include all necessary files and dependencies
- **Documentation** - Provide clear setup and usage instructions
- **Error Handling** - Show proper error handling patterns
- **Comments** - Explain key concepts and gotchas
- **Testing** - Include instructions for testing/validation

## Support

If you encounter issues with any examples:

1. Check the example's individual documentation
2. Verify your MSFS and SimConnect setup
3. Review the [Troubleshooting Guide](../advanced/troubleshooting.md)
4. Open an issue on GitHub with specific error details

## Next Steps

After working through the examples, explore:
- [API Reference](../api/) - Detailed API documentation
- [Advanced Topics](../advanced/) - Performance optimization, architecture patterns
- [Contributing Guide](../../CONTRIBUTING.md) - Help improve the library
