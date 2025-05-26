# Performance Optimization

This guide covers performance optimization techniques for go-simconnect applications, including data collection strategies, memory management, and production deployment considerations.

## Data Collection Performance

### Update Rate Optimization

**Recommended Settings:**
```go
// Use SIMCONNECT_PERIOD_SECOND with CHANGED flag for optimal performance
// This provides ~20Hz update rate when data actually changes
fdm := client.NewFlightDataManager(simClient)
fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
fdm.Start() // Uses PERIOD_SECOND + DATA_REQUEST_FLAG_CHANGED by default
```

**Why This Works:**
- `SIMCONNECT_PERIOD_SECOND` provides consistent timing
- `SIMCONNECT_DATA_REQUEST_FLAG_CHANGED` only sends updates when values change
- Reduces unnecessary network traffic and CPU usage
- Ideal for real-time applications that need responsive updates

### Individual Data Definitions

**✅ Recommended Approach:**
```go
// Each variable gets its own data definition
fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
fdm.AddVariable("Altitude", "INDICATED ALTITUDE", "feet")
fdm.AddVariable("Heading", "PLANE HEADING DEGREES MAGNETIC", "degrees")
```

**❌ Avoid Combined Definitions:**
```go
// Don't try to combine multiple variables into single definition
// This can cause SimConnect exceptions and performance issues
```

**Benefits:**
- Prevents SimConnect exceptions from invalid variable combinations
- Allows individual error handling per variable
- Better performance with selective updates
- Easier debugging and maintenance

## Memory Management

### Return Copies, Not References

**✅ Thread-Safe Pattern:**
```go
func (fdm *FlightDataManager) GetAllVariables() map[string]FlightVariable {
    fdm.mutex.RLock()
    defer fdm.mutex.RUnlock()
    
    // Return a copy to prevent external modification
    result := make(map[string]FlightVariable)
    for k, v := range fdm.variables {
        result[k] = v // Copy the struct
    }
    return result
}
```

**❌ Dangerous Direct Access:**
```go
// Never return direct references to internal data
func (fdm *FlightDataManager) GetVariables() *map[string]FlightVariable {
    return &fdm.variables // Allows external modification!
}
```

### Resource Cleanup

**Always Use Defer:**
```go
func connectAndProcess() error {
    simClient := client.NewClient("MyApp")
    if err := simClient.Open(); err != nil {
        return err
    }
    defer simClient.Close() // Ensures cleanup even on panic
    
    fdm := client.NewFlightDataManager(simClient)
    if err := fdm.Start(); err != nil {
        return err
    }
    defer fdm.Stop() // Cleanup data collection
    
    // ... application logic
    return nil
}
```

### Goroutine Management

**Monitor Long-Running Goroutines:**
```go
func monitorDataManager(fdm *FlightDataManager, ctx context.Context) {
    // Use context for graceful shutdown
    errorChan := fdm.GetErrors()
    
    for {
        select {
        case err := <-errorChan:
            log.Printf("Data error: %v", err)
        case <-ctx.Done():
            log.Println("Shutting down data monitoring")
            return
        }
    }
}
```

## Network and SimConnect Optimization

### Connection Resilience

**Implement Retry Logic:**
```go
func connectWithRetry(appName string, maxRetries int) (*client.Client, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        simClient := client.NewClient(appName)
        if err := simClient.Open(); err != nil {
            lastErr = err
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }
        return simClient, nil
    }
    
    return nil, fmt.Errorf("failed to connect after %d retries: %v", maxRetries, lastErr)
}
```

### Error Channel Buffering

**Use Buffered Channels:**
```go
// FlightDataManager uses buffered error channels to prevent blocking
errorChan := make(chan error, 10) // Buffer capacity of 10

// This prevents blocking when errors occur faster than they're consumed
```

**Monitor Error Channel:**
```go
go func() {
    for err := range fdm.GetErrors() {
        // Process errors in separate goroutine to prevent blocking
        log.Printf("SimConnect error: %v", err)
        
        // Implement error recovery logic here
        if strings.Contains(err.Error(), "connection lost") {
            // Attempt reconnection
        }
    }
}()
```

## Production Deployment

### Application Lifecycle

**Graceful Startup:**
```go
func main() {
    // Check SimConnect availability before starting
    if !isSimConnectAvailable() {
        log.Println("SimConnect not available, running in offline mode")
        return
    }
    
    // Setup signal handling for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        log.Println("Shutting down gracefully...")
        cancel()
    }()
    
    // Start application with context
    if err := runApplication(ctx); err != nil {
        log.Fatalf("Application error: %v", err)
    }
}
```

### Performance Monitoring

**Track Key Metrics:**
```go
type PerformanceMonitor struct {
    dataUpdatesPerSecond atomic.Int64
    errorCount          atomic.Int64
    lastUpdateTime      atomic.Value // time.Time
}

func (pm *PerformanceMonitor) RecordDataUpdate() {
    pm.dataUpdatesPerSecond.Add(1)
    pm.lastUpdateTime.Store(time.Now())
}

func (pm *PerformanceMonitor) GetStats() (updatesPerSec int64, errors int64, lastUpdate time.Time) {
    return pm.dataUpdatesPerSecond.Load(), 
           pm.errorCount.Load(), 
           pm.lastUpdateTime.Load().(time.Time)
}
```

## Benchmarking

### Data Collection Performance

**Typical Performance Characteristics:**
- **Update Rate**: 20-60 Hz depending on MSFS load
- **Memory Usage**: ~1-2MB for 50 variables
- **CPU Usage**: <1% on modern systems
- **Network Latency**: <1ms for local SimConnect

**Performance Testing:**
```go
func BenchmarkDataCollection(b *testing.B) {
    simClient := client.NewClient("BenchmarkApp")
    if err := simClient.Open(); err != nil {
        b.Skip("SimConnect not available")
    }
    defer simClient.Close()
    
    fdm := client.NewFlightDataManager(simClient)
    fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
    fdm.Start()
    defer fdm.Stop()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = fdm.GetVariable("Airspeed")
    }
}
```

## Common Performance Issues

### Issue: High CPU Usage

**Symptoms:**
- CPU usage >5% when idle
- Slow response to variable requests

**Solutions:**
1. Check update rate - avoid SIMCONNECT_PERIOD_VISUAL_FRAME
2. Reduce number of variables being tracked
3. Use DATA_REQUEST_FLAG_CHANGED to reduce updates

### Issue: Memory Leaks

**Symptoms:**
- Gradual memory increase over time
- Application becomes slower

**Solutions:**
1. Ensure proper cleanup with defer statements
2. Don't store references to returned data structures
3. Monitor goroutine count for leaks

### Issue: Connection Timeouts

**Symptoms:**
- Frequent connection errors
- "SimConnect not responding" messages

**Solutions:**
1. Implement connection retry logic
2. Check MSFS load and performance
3. Reduce data request frequency if MSFS is under load

## Best Practices Summary

1. **Use individual data definitions** for each variable
2. **Implement retry logic** for production reliability
3. **Monitor error channels** in separate goroutines
4. **Use buffered channels** to prevent blocking
5. **Return copies** of data structures, not references
6. **Implement graceful shutdown** with context cancellation
7. **Track performance metrics** for production monitoring
8. **Test with realistic loads** before deployment

## Related Documentation

- [Architecture Patterns](architecture.md) - Design patterns for robust applications
- [Troubleshooting Guide](troubleshooting.md) - Common issues and solutions
- [API Reference](../api/) - Complete API documentation
