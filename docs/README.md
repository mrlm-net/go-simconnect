# go-simconnect Documentation

Welcome to the complete documentation for the go-simconnect package - a production-ready Go library for Microsoft Flight Simulator 2024 SimConnect integration.

## Quick Navigation

### Getting Started
- [Installation & Setup](getting-started.md)
- [Quick Start Guide](getting-started.md#quick-start)
- [Basic Examples](../examples/)

### API Documentation
- [SimConnect Client](api/client.md) - Core SimConnect connection management
- [Flight Data Manager](api/flight-data-manager.md) - Real-time data collection and control
- [Available Variables](api/variables.md) - Complete reference of simulation variables
- [Error Handling](api/errors.md) - Error types and handling strategies

### Examples & Guides
- [Basic Usage](examples/basic-usage.md) - Simple data reading
- [Camera Control](examples/camera-control.md) - SetData functionality demonstration
- [Web Dashboard](examples/web-dashboard.md) - Real-time web interface
- [Production Integration](examples/production.md) - Best practices for production use

### Advanced Topics
- [Performance Optimization](advanced/performance.md)
- [Troubleshooting Guide](advanced/troubleshooting.md)
- [Architecture Deep Dive](advanced/architecture.md)
- [Contributing](../CONTRIBUTING.md)

## Package Overview

go-simconnect provides a high-level, thread-safe interface for:

- **Real-time flight data collection** (~20Hz update rate)
- **Aircraft control** via SetData functionality  
- **Camera and view management**
- **System state monitoring**
- **Extensible variable system**

## Key Features

✅ **Production Ready** - Comprehensive error handling and thread safety  
✅ **Type Safe** - Strong typing for all simulation variables  
✅ **High Performance** - Optimized data collection with minimal overhead  
✅ **Extensible** - Easy to add custom variables and functionality  
✅ **Well Documented** - Complete API documentation and examples  

## Recent Updates

- **SetData Support** - Full read/write capability for simulation variables
- **Camera Control** - Dynamic camera view switching
- **Enhanced Examples** - More comprehensive demonstration applications
- **Improved Documentation** - Restructured for better navigation

## Quick Example

```go
package main

import (
    "log"
    "github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() {
    // Connect to simulator
    simClient := client.NewClient("MyApp")
    if err := simClient.Open(); err != nil {
        log.Fatal(err)
    }
    defer simClient.Close()

    // Create flight data manager
    fdm := client.NewFlightDataManager(simClient)
    
    // Add variables to monitor
    fdm.AddVariable("Altitude", "Plane Altitude", "feet")
    fdm.AddVariable("Speed", "Airspeed Indicated", "knots")
    
    // Start data collection
    fdm.Start()
    defer fdm.Stop()
    
    // Read data
    if altitude, found := fdm.GetVariable("Altitude"); found {
        log.Printf("Current altitude: %.0f feet", altitude.Value)
    }
}
```

## Support

- **Issues**: [GitHub Issues](https://github.com/mrlm-net/go-simconnect/issues)
- **Examples**: See `examples/` directory
- **API Reference**: See `docs/api/` directory

---

For detailed installation and setup instructions, start with the [Getting Started Guide](getting-started.md).
