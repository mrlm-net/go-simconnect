# mrlm-net/go-simconnect

Production-ready Go package for Microsoft Flight Simulator 2024 SimConnect integration, providing real-time flight data access and aircraft control.

|  |  |
|---|---|
| **Package name** | github.com/mrlm-net/go-simconnect |
| **Package version** | ![GitHub Release](https://img.shields.io/github/v/release/mrlm-net/go-simconnect) |
| **Latest version** | ![GitHub Release](https://img.shields.io/github/v/release/mrlm-net/go-simconnect) |
| **License** | ![GitHub License](https://img.shields.io/github/license/mrlm-net/go-simconnect) |

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
    // Create and connect to SimConnect
    client, err := simconnect.NewClient("MyFlightApp")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }

    // Create flight data manager
    fdm := client.NewFlightDataManager(client)

    // Add variables to track
    fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
    fdm.AddVariable("Altitude", "INDICATED ALTITUDE", "feet")
    fdm.AddVariableWithWritable("Camera", "CAMERA STATE", "number", true)

    // Start data collection
    if err := fdm.Start(); err != nil {
        log.Fatal(err)
    }
    defer fdm.Stop()

    // Read and control simulation data
    for i := 0; i < 10; i++ {
        if variable, found := fdm.GetVariable("Airspeed"); found {
            fmt.Printf("Airspeed: %.1f knots\n", variable.Value)
        }
        
        // Change camera view
        fdm.SetVariable("Camera", float64(2+i%4))
        
        time.Sleep(2 * time.Second)
    }
}
```

## Features

- ✅ **Real-time Flight Data** - Position, speed, attitude, engine parameters
- ✅ **Aircraft Control** - Set variables, control systems, change camera views
- ✅ **Thread-safe Operations** - Concurrent access with proper synchronization
- ✅ **Comprehensive API** - Full SimConnect variable access with 200+ documented variables
- ✅ **Production Ready** - Error handling, statistics, and performance optimization
- ✅ **Rich Examples** - Web dashboard, camera control, complete demos

## Documentation

### 📚 [Getting Started Guide](docs/getting-started.md)
Installation, setup, and your first SimConnect application.

### 📖 [API Reference](docs/api/)
- [Client API](docs/api/client.md) - Core SimConnect client functionality
- [FlightDataManager](docs/api/flight-data-manager.md) - High-level data management
- [Variables Reference](docs/api/variables.md) - 200+ available SimConnect variables

### 💡 [Examples](docs/examples/)
- [Camera Control](examples/camera_test/) - Real-time camera view switching
- [Web Dashboard](examples/web_dashboard/) - Browser-based flight data display
- [Complete Demo](examples/final_complete_demo_fixed/) - Comprehensive feature showcase

### 🔧 [Advanced Topics](docs/advanced/)
- [Performance Optimization](docs/advanced/performance.md)
- [Troubleshooting Guide](docs/advanced/troubleshooting.md)
- [Architecture Patterns](docs/advanced/architecture.md)

## Installation

```bash
go get github.com/mrlm-net/go-simconnect
```

**Requirements:** Microsoft Flight Simulator 2024, Windows OS, Go 1.19+

## Examples

### 🎥 Camera Control
Test SetData functionality with immediate visual feedback:
```bash
cd examples/camera_test
go run main.go
```

### 🌐 Web Dashboard  
Modern web interface for flight data:
```bash
cd examples/web_dashboard
go run main.go
# Open http://localhost:8080
```

### 🛠️ Complete Demo
Comprehensive feature showcase:
```bash
cd examples/final_complete_demo_fixed
go run main.go
## Support

**Issues & Questions:** [GitHub Issues](https://github.com/mrlm-net/go-simconnect/issues)  
**Troubleshooting:** [Troubleshooting Guide](docs/advanced/troubleshooting.md)  
**API Reference:** [Complete API Documentation](docs/api/)

## Contributing

Contributions welcome! See our [Contributing Guidelines](CONTRIBUTING.md).

### Development
- Follow standard Go conventions
- Maintain thread-safety for all public APIs
- Include comprehensive error handling
- Write tests for new functionality

## License

Licensed under the MIT License. See [LICENSE](LICENSE) file for details.

---

**go-simconnect** - Production-ready SimConnect integration for Go developers  
© 2024 Martin Hrášek & WANTED.solutions s.r.o.