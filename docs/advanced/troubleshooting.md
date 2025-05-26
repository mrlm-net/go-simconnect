# Troubleshooting Guide

This guide covers common issues encountered when using go-simconnect and their solutions.

## Connection Issues

### "Failed to create SimConnect client"

**Symptoms:**
- Application fails to start
- Error message about SimConnect client creation

**Causes and Solutions:**

1. **SimConnect DLL not found**
   ```
   Solution: Ensure SimConnect.dll is in your system PATH or application directory
   Location: Usually in MSFS installation directory under SDK/Core Utilities Kit/lib
   ```

2. **Insufficient permissions**
   ```
   Solution: Run application as administrator or ensure proper user permissions
   ```

3. **MSFS not running**
   ```
   Solution: Start Microsoft Flight Simulator before running your application
   ```

### "Connection refused" or "Cannot connect to SimConnect"

**Symptoms:**
- Client created successfully but connection fails
- Timeout errors during connection

**Causes and Solutions:**

1. **MSFS SimConnect disabled**
   ```
   Solution: Enable SimConnect in MSFS Developer Mode settings
   Path: Options → General → Developers → Enable Development Mode
   ```

2. **Network firewall blocking connection**
   ```
   Solution: Add firewall exception for your application and MSFS
   ```

3. **MSFS not fully loaded**
   ```
   Solution: Wait for MSFS to fully load an aircraft before connecting
   ```

## Data Collection Issues

### "No data received" or "Variables not updating"

**Symptoms:**
- FlightDataManager starts but no data received
- Variable values remain unchanged

**Causes and Solutions:**

1. **Variables not added before starting**
   ```go
   // Wrong:
   fdm.Start()
   fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
   
   // Correct:
   fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
   fdm.Start()
   ```

2. **Invalid variable names or units**
   ```go
   // Wrong:
   fdm.AddVariable("Speed", "SPEED", "mph")
   
   // Correct:
   fdm.AddVariable("Speed", "AIRSPEED INDICATED", "knots")
   ```

3. **Aircraft doesn't support requested variables**
   ```
   Solution: Check if the aircraft supports the specific systems/variables
   Test with different aircraft (e.g., Cessna 152 vs. Airbus A320)
   ```

### "Variable not found" errors

**Symptoms:**
- Specific variables fail to be added
- Error messages about unknown variables

**Causes and Solutions:**

1. **Incorrect variable name**
   ```
   Solution: Verify variable names against SimConnect SDK documentation
   Common mistakes: "ALTITUDE" vs "INDICATED ALTITUDE"
   ```

2. **Incorrect units**
   ```
   Solution: Use exact unit strings from SimConnect documentation
   Common mistakes: "ft" vs "feet", "kts" vs "knots"
   ```

3. **Case sensitivity**
   ```
   Solution: Variable names and units are case-sensitive
   Use exact capitalization as documented
   ```

## Performance Issues

### High CPU usage

**Symptoms:**
- Application consuming excessive CPU
- System becomes unresponsive

**Causes and Solutions:**

1. **Too many variables or high update rate**
   ```go
   // Reduce number of variables or use change detection
   fdm.AddVariable("Essential1", "AIRSPEED INDICATED", "knots")
   // Only add variables you actually need
   ```

2. **Inefficient data processing**
   ```go
   // Use goroutines for heavy processing
   go func() {
       for {
           vars := fdm.GetAllVariables()
           // Process data in background
       }
   }()
   ```

3. **Memory leaks in loops**
   ```go
   // Avoid creating objects in tight loops
   ticker := time.NewTicker(time.Second)
   defer ticker.Stop()
   
   for range ticker.C {
       // Process data
   }
   ```

### Memory leaks

**Symptoms:**
- Memory usage continuously increasing
- Application eventually crashes

**Causes and Solutions:**

1. **Unclosed channels or goroutines**
   ```go
   // Always close channels and stop goroutines
   defer fdm.Stop()
   defer client.Close()
   ```

2. **Error channel not drained**
   ```go
   // Drain error channel to prevent backup
   go func() {
       for err := range fdm.GetErrors() {
           log.Printf("Error: %v", err)
       }
   }()
   ```

## SetData/Write Operations Issues

### "Cannot set variable" errors

**Symptoms:**
- SetVariable calls fail
- No effect in simulator

**Causes and Solutions:**

1. **Variable not marked as writable**
   ```go
   // Wrong:
   fdm.AddVariable("AP Master", "AUTOPILOT MASTER", "bool")
   
   // Correct:
   fdm.AddVariableWithWritable("AP Master", "AUTOPILOT MASTER", "bool", true)
   ```

2. **Invalid value range**
   ```go
   // Check valid ranges for variables
   // Example: Camera state values 2-6, not 0-10
   fdm.SetVariable("Camera", 2) // Valid
   fdm.SetVariable("Camera", 99) // Invalid
   ```

3. **Aircraft system not ready**
   ```
   Solution: Wait for aircraft systems to initialize
   Check system state before setting values
   ```

## Build and Compilation Issues

### "Cannot find package" errors

**Symptoms:**
- Import errors during build
- Package not found messages

**Causes and Solutions:**

1. **Module not initialized**
   ```bash
   go mod init your-app-name
   go mod tidy
   ```

2. **Incorrect import path**
   ```go
   // Correct import:
   import "github.com/mrlm-net/go-simconnect/pkg/client"
   ```

3. **Dependency not downloaded**
   ```bash
   go get github.com/mrlm-net/go-simconnect
   ```

### "DLL not found" runtime errors

**Symptoms:**
- Application builds but fails to run
- Missing DLL errors

**Causes and Solutions:**

1. **SimConnect.dll not in PATH**
   ```
   Solution: Copy SimConnect.dll to application directory
   Or add MSFS SDK directory to system PATH
   ```

2. **Architecture mismatch**
   ```
   Solution: Ensure 64-bit DLL for 64-bit application
   Use correct SimConnect SDK version
   ```

## Debugging Techniques

### Enable Debug Logging

```go
// Add debug logging to your application
log.SetLevel(log.DebugLevel)

// Monitor connection state
if client.IsConnected() {
    log.Debug("SimConnect connected successfully")
} else {
    log.Debug("SimConnect not connected")
}
```

### Monitor Data Flow

```go
// Add statistics monitoring
dataCount, errorCount, lastUpdate := fdm.GetStats()
log.Printf("Data: %d, Errors: %d, Last Update: %v", 
    dataCount, errorCount, lastUpdate)
```

### Test with Minimal Example

When troubleshooting, start with a minimal working example:

```go
package main

import (
    "log"
    "time"
    "github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() {
    // Minimal test
    client, err := simconnect.NewClient("TestApp")
    if err != nil {
        log.Fatal("Client creation failed:", err)
    }
    defer client.Close()

    if err := client.Connect(); err != nil {
        log.Fatal("Connection failed:", err)
    }

    log.Println("Connection successful!")
    time.Sleep(5 * time.Second)
}
```

## Getting Help

If you're still experiencing issues:

1. **Check MSFS Version** - Ensure compatibility with your MSFS version
2. **Review Examples** - Compare your code with working examples
3. **Enable Verbose Logging** - Add detailed logging to identify the issue
4. **Test Different Aircraft** - Some variables are aircraft-specific
5. **Check GitHub Issues** - Search for similar reported issues
6. **Create Minimal Reproduction** - Isolate the problem in a simple test case

## Reporting Bugs

When reporting issues, please include:

- Go version (`go version`)
- MSFS version and edition
- Complete error messages
- Minimal code that reproduces the issue
- System information (Windows version, architecture)
- Steps to reproduce the problem

## Common Error Messages and Solutions

| Error Message | Likely Cause | Solution |
|---------------|--------------|----------|
| "SimConnect not available" | MSFS not running | Start MSFS first |
| "Object name already exists" | Duplicate client name | Use unique client names |
| "Data request failed" | Invalid variable/units | Check variable names and units |
| "Connection lost" | MSFS closed/crashed | Implement reconnection logic |
| "Access denied" | Permissions issue | Run as administrator |
| "Timeout" | Network/firewall issue | Check firewall settings |
