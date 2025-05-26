# Getting Started with go-simconnect

## Installation

### Prerequisites

- **Microsoft Flight Simulator 2024** installed and running
- **Windows Operating System** (SimConnect is Windows-only)
- **Go 1.19 or later**

### Install Package

```bash
go get github.com/mrlm-net/go-simconnect
```

### SimConnect SDK

The package automatically locates SimConnect.dll in common installation paths:
- `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
- `C:\MSFS SDK\SimConnect SDK\lib\SimConnect.dll`

If you have a custom installation, specify the path manually:

```go
client := client.NewClientWithDLLPath("MyApp", "C:\\Custom\\Path\\SimConnect.dll")
```

## Quick Start

### Basic Connection

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() {
    // Create client
    simClient := client.NewClient("QuickStart")
    
    // Connect to simulator
    if err := simClient.Open(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer simClient.Close()
    
    fmt.Println("✅ Connected to Microsoft Flight Simulator 2024!")
}
```

### Reading Flight Data

```go
func main() {
    // ... connection code ...
    
    // Create flight data manager
    fdm := client.NewFlightDataManager(simClient)
    
    // Add variables to monitor
    fdm.AddVariable("Altitude", "Plane Altitude", "feet")
    fdm.AddVariable("Speed", "Airspeed Indicated", "knots")
    fdm.AddVariable("Heading", "Plane Heading Degrees Magnetic", "degrees")
    
    // Start data collection
    if err := fdm.Start(); err != nil {
        log.Fatalf("Failed to start data collection: %v", err)
    }
    defer fdm.Stop()
    
    // Wait for data to be collected
    time.Sleep(2 * time.Second)
    
    // Read current values
    if altitude, found := fdm.GetVariable("Altitude"); found {
        fmt.Printf("Altitude: %.0f feet\n", altitude.Value)
    }
    
    if speed, found := fdm.GetVariable("Speed"); found {
        fmt.Printf("Speed: %.0f knots\n", speed.Value)
    }
    
    if heading, found := fdm.GetVariable("Heading"); found {
        fmt.Printf("Heading: %.0f degrees\n", heading.Value)
    }
}
```

### Writing Data (SetData)

```go
func main() {
    // ... connection and setup code ...
    
    // Add writable variable
    fdm.AddVariableWithWritable("Camera State", "Camera State", "number", true)
    fdm.Start()
    defer fdm.Stop()
    
    // Wait for initial data
    time.Sleep(2 * time.Second)
    
    // Change camera to external view
    if err := fdm.SetVariable("Camera State", 3.0); err != nil {
        log.Printf("Failed to set camera: %v", err)
    } else {
        fmt.Println("✅ Camera changed to external view!")
    }
}
```

## Common Variables

Here are some frequently used simulation variables:

### Position & Navigation
```go
fdm.AddVariable("Altitude", "Plane Altitude", "feet")
fdm.AddVariable("Latitude", "Plane Latitude", "degrees")
fdm.AddVariable("Longitude", "Plane Longitude", "degrees")
fdm.AddVariable("Heading", "Plane Heading Degrees Magnetic", "degrees")
fdm.AddVariable("Ground Speed", "Ground Velocity", "knots")
```

### Aircraft State
```go
fdm.AddVariable("Indicated Airspeed", "Airspeed Indicated", "knots")
fdm.AddVariable("Vertical Speed", "Vertical Speed", "feet per minute")
fdm.AddVariable("Bank Angle", "Plane Bank Degrees", "degrees")
fdm.AddVariable("Pitch Angle", "Plane Pitch Degrees", "degrees")
```

### Engine & Controls (Writable)
```go
fdm.AddVariableWithWritable("Throttle", "General Eng Throttle Lever Position:1", "percent", true)
fdm.AddVariableWithWritable("Flaps", "Flaps Handle Percent", "percent", true)
fdm.AddVariableWithWritable("Gear", "Gear Handle Position", "bool", true)
```

## Error Handling

Always implement proper error handling:

```go
// Connection errors
if err := simClient.Open(); err != nil {
    log.Fatalf("Connection failed: %v", err)
}

// Variable addition errors
if err := fdm.AddVariable("Invalid", "NonExistent", "units"); err != nil {
    log.Printf("Warning: Could not add variable: %v", err)
}

// Data collection errors
go func() {
    for err := range fdm.GetErrors() {
        log.Printf("Data error: %v", err)
    }
}()
```

## Next Steps

Now that you have the basics working, explore these features:

### Examples
- **[Web Dashboard](examples/web-dashboard.md)** - Browser-based flight data display
- **[Camera Control](examples/camera-test.md)** - Real-time camera switching
- **[System Events](examples/system-events.md)** - Event-driven notifications
- **[Complete Demo](examples/complete-demo.md)** - Advanced patterns and best practices

### Core Features
- **[FlightDataManager](api/flight-data-manager.md)** - High-level data management
- **[SystemEventManager](api/system-events.md)** - Event-driven programming
- **[Client API](api/client.md)** - Low-level SimConnect access

### Advanced Topics
- **[Performance Optimization](advanced/performance.md)** - Scaling and efficiency
- **[Architecture Patterns](advanced/architecture.md)** - Production-ready designs
- **[Troubleshooting](advanced/troubleshooting.md)** - Common issues and solutions

## Troubleshooting

If you encounter issues:

1. **Ensure MSFS 2024 is running** before connecting
2. **Check SimConnect.dll path** - use custom path if needed
3. **Verify variable names** - use exact SimConnect variable names
4. **Handle connection failures** - implement retry logic for production

For detailed troubleshooting, see [Troubleshooting Guide](advanced/troubleshooting.md).
