# Architecture Patterns

This guide covers design patterns and architectural approaches for building robust, maintainable applications with go-simconnect.

## Core Architecture Principles

### 1. Thread-Safe Design

All public APIs in go-simconnect are designed to be thread-safe from the ground up:

```go
type FlightDataManager struct {
    mutex     sync.RWMutex          // Protects all internal state
    variables map[string]FlightVariable
    client    *Client
    running   bool
    errors    chan error
}

// All public methods use proper locking
func (fdm *FlightDataManager) GetVariable(name string) (FlightVariable, bool) {
    fdm.mutex.RLock()
    defer fdm.mutex.RUnlock()
    
    variable, exists := fdm.variables[name]
    return variable, exists
}
```

**Key Principles:**
- Use `sync.RWMutex` for read-heavy operations
- Always use `defer` for unlock operations
- Return copies of data structures, never references
- Separate read and write operations clearly

### 2. Individual Data Definitions

Each SimConnect variable gets its own data definition to prevent exceptions and improve reliability:

```go
// Each variable is independent
type VariableDefinition struct {
    DataDefinitionID      uint32
    SimObjectDataRequestID uint32
    SimVarName           string
    Units                string
    Writable             bool
}

// This prevents SimConnect exceptions from incompatible variable combinations
```

**Benefits:**
- Isolation: One variable failure doesn't affect others
- Flexibility: Can add/remove variables independently
- Debugging: Easier to trace issues to specific variables
- Performance: SimConnect handles individual definitions efficiently

## Application Architecture Patterns

### 1. Layered Architecture

**Recommended Application Structure:**
```
Application Layer (main.go, handlers)
    ↓
Service Layer (business logic)
    ↓
Data Access Layer (FlightDataManager)
    ↓
SimConnect Layer (Client)
    ↓
Microsoft Flight Simulator
```

**Example Implementation:**
```go
// Service Layer
type FlightService struct {
    dataManager *client.FlightDataManager
    mu          sync.RWMutex
}

func (fs *FlightService) GetFlightStatus() FlightStatus {
    fs.mu.RLock()
    defer fs.mu.RUnlock()
    
    airspeed, _ := fs.dataManager.GetVariable("Airspeed")
    altitude, _ := fs.dataManager.GetVariable("Altitude")
    
    return FlightStatus{
        Airspeed: airspeed.Value,
        Altitude: altitude.Value,
        Status:   fs.calculateStatus(airspeed.Value, altitude.Value),
    }
}

// Application Layer
func main() {
    // Initialize layers
    simClient := client.NewClient("MyApp")
    dataManager := client.NewFlightDataManager(simClient)
    flightService := NewFlightService(dataManager)
    
    // Setup and run
    setupVariables(dataManager)
    startServices(simClient, dataManager, flightService)
}
```

### 2. Event-Driven Architecture

**For Real-Time Applications:**
```go
type FlightEventBus struct {
    subscribers map[string][]chan FlightEvent
    mu          sync.RWMutex
}

type FlightEvent struct {
    Type      string
    Variable  string
    Value     float64
    Timestamp time.Time
}

func (bus *FlightEventBus) Subscribe(eventType string) <-chan FlightEvent {
    bus.mu.Lock()
    defer bus.mu.Unlock()
    
    eventChan := make(chan FlightEvent, 100) // Buffered
    bus.subscribers[eventType] = append(bus.subscribers[eventType], eventChan)
    return eventChan
}

func (bus *FlightEventBus) Publish(event FlightEvent) {
    bus.mu.RLock()
    defer bus.mu.RUnlock()
    
    for _, subscriber := range bus.subscribers[event.Type] {
        select {
        case subscriber <- event:
        default:
            // Subscriber channel full, skip to prevent blocking
        }
    }
}
```

### 3. Repository Pattern

**For Data Abstraction:**
```go
type FlightDataRepository interface {
    GetAirspeed() (float64, error)
    GetAltitude() (float64, error)
    SetCameraView(view int) error
    IsConnected() bool
}

type SimConnectRepository struct {
    dataManager *client.FlightDataManager
}

func (r *SimConnectRepository) GetAirspeed() (float64, error) {
    if variable, found := r.dataManager.GetVariable("Airspeed"); found {
        return variable.Value, nil
    }
    return 0, errors.New("airspeed data not available")
}

// This allows for easy testing with mock implementations
type MockRepository struct {
    airspeed float64
    altitude float64
}

func (m *MockRepository) GetAirspeed() (float64, error) {
    return m.airspeed, nil
}
```

## Error Handling Patterns

### 1. Centralized Error Management

**Error Aggregation:**
```go
type ErrorManager struct {
    errors    chan error
    handlers  []ErrorHandler
    mu        sync.RWMutex
    running   bool
}

type ErrorHandler interface {
    HandleError(err error) bool // Returns true if error was handled
}

func (em *ErrorManager) Start() {
    em.mu.Lock()
    em.running = true
    em.mu.Unlock()
    
    go func() {
        for err := range em.errors {
            em.handleError(err)
        }
    }()
}

func (em *ErrorManager) handleError(err error) {
    em.mu.RLock()
    handlers := em.handlers
    em.mu.RUnlock()
    
    for _, handler := range handlers {
        if handler.HandleError(err) {
            return // Error was handled
        }
    }
    
    // Unhandled error
    log.Printf("Unhandled error: %v", err)
}
```

### 2. Circuit Breaker Pattern

**For Connection Resilience:**
```go
type CircuitBreaker struct {
    failureCount    int
    lastFailureTime time.Time
    state          CircuitState
    threshold      int
    timeout        time.Duration
    mu             sync.RWMutex
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if cb.state == Open {
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = HalfOpen
        } else {
            return errors.New("circuit breaker is open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()
        
        if cb.failureCount >= cb.threshold {
            cb.state = Open
        }
        return err
    }
    
    // Success - reset circuit breaker
    cb.failureCount = 0
    cb.state = Closed
    return nil
}
```

## Configuration Patterns

### 1. Configuration Structure

**Centralized Configuration:**
```go
type Config struct {
    SimConnect SimConnectConfig `json:"simconnect"`
    App        AppConfig        `json:"app"`
    Variables  []VariableConfig `json:"variables"`
}

type SimConnectConfig struct {
    AppName         string `json:"app_name"`
    DLLPath         string `json:"dll_path,omitempty"`
    RetryAttempts   int    `json:"retry_attempts"`
    RetryDelay      string `json:"retry_delay"`
}

type VariableConfig struct {
    Name      string `json:"name"`
    SimVar    string `json:"sim_var"`
    Units     string `json:"units"`
    Writable  bool   `json:"writable"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### 2. Environment-Based Configuration

**Development vs Production:**
```go
type Environment string

const (
    Development Environment = "development"
    Production  Environment = "production"
    Testing     Environment = "testing"
)

func GetConfig(env Environment) Config {
    switch env {
    case Development:
        return Config{
            SimConnect: SimConnectConfig{
                AppName:       "DevApp",
                RetryAttempts: 3,
                RetryDelay:    "1s",
            },
            // More permissive error handling
        }
    case Production:
        return Config{
            SimConnect: SimConnectConfig{
                AppName:       "ProductionApp",
                RetryAttempts: 10,
                RetryDelay:    "5s",
            },
            // Robust error handling and monitoring
        }
    default:
        return getDefaultConfig()
    }
}
```

## Testing Patterns

### 1. Dependency Injection

**Testable Architecture:**
```go
type Dependencies struct {
    SimClient   SimConnectClient
    DataManager DataManager
    Logger      Logger
}

type SimConnectClient interface {
    Open() error
    Close() error
    IsConnected() bool
}

type DataManager interface {
    AddVariable(name, simVar, units string) error
    GetVariable(name string) (FlightVariable, bool)
    Start() error
    Stop() error
}

type Application struct {
    deps Dependencies
}

func (app *Application) Run() error {
    if err := app.deps.SimClient.Open(); err != nil {
        return err
    }
    defer app.deps.SimClient.Close()
    
    // Application logic using interfaces
    return nil
}
```

### 2. Mock Implementations

**For Unit Testing:**
```go
type MockDataManager struct {
    variables map[string]FlightVariable
    errors    []error
    started   bool
}

func (m *MockDataManager) AddVariable(name, simVar, units string) error {
    if len(m.errors) > 0 {
        err := m.errors[0]
        m.errors = m.errors[1:]
        return err
    }
    
    m.variables[name] = FlightVariable{
        Name:  name,
        Value: 100.0, // Mock value
        Units: units,
    }
    return nil
}

func (m *MockDataManager) GetVariable(name string) (FlightVariable, bool) {
    variable, exists := m.variables[name]
    return variable, exists
}

// Use in tests
func TestFlightService(t *testing.T) {
    mockDataManager := &MockDataManager{
        variables: make(map[string]FlightVariable),
    }
    
    service := NewFlightService(mockDataManager)
    // Test without needing actual SimConnect
}
```

## Deployment Patterns

### 1. Health Check Pattern

**Service Health Monitoring:**
```go
type HealthChecker struct {
    checks []HealthCheck
    mu     sync.RWMutex
}

type HealthCheck interface {
    Name() string
    Check() error
}

type SimConnectHealthCheck struct {
    client *client.Client
}

func (sc *SimConnectHealthCheck) Name() string {
    return "SimConnect"
}

func (sc *SimConnectHealthCheck) Check() error {
    if !sc.client.IsConnected() {
        return errors.New("SimConnect not connected")
    }
    return nil
}

func (hc *HealthChecker) CheckAll() map[string]error {
    hc.mu.RLock()
    defer hc.mu.RUnlock()
    
    results := make(map[string]error)
    for _, check := range hc.checks {
        results[check.Name()] = check.Check()
    }
    return results
}
```

### 2. Graceful Shutdown Pattern

**Clean Application Termination:**
```go
type Application struct {
    simClient   *client.Client
    dataManager *client.FlightDataManager
    webServer   *http.Server
    done        chan struct{}
}

func (app *Application) Start() error {
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    // Start components
    if err := app.startComponents(); err != nil {
        return err
    }
    
    // Wait for shutdown signal
    go func() {
        <-sigChan
        log.Println("Shutdown signal received")
        app.shutdown()
    }()
    
    <-app.done
    return nil
}

func (app *Application) shutdown() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Shutdown in reverse order
    if app.webServer != nil {
        app.webServer.Shutdown(ctx)
    }
    
    if app.dataManager != nil {
        app.dataManager.Stop()
    }
    
    if app.simClient != nil {
        app.simClient.Close()
    }
    
    close(app.done)
}
```

## Best Practices Summary

### Design Principles
1. **Thread Safety First** - Use proper synchronization for all shared state
2. **Fail Fast** - Validate inputs and fail early with clear error messages
3. **Separation of Concerns** - Keep SimConnect logic separate from business logic
4. **Interface Segregation** - Define small, focused interfaces for testability
5. **Dependency Injection** - Make dependencies explicit for testing and flexibility

### Error Handling
1. **Centralized Error Management** - Aggregate and handle errors consistently
2. **Circuit Breaker** - Protect against cascading failures
3. **Graceful Degradation** - Continue operating with reduced functionality when possible
4. **Comprehensive Logging** - Log errors with context for debugging

### Performance
1. **Resource Cleanup** - Always use defer for cleanup operations
2. **Buffered Channels** - Prevent blocking with appropriate buffer sizes
3. **Individual Data Definitions** - Use separate definitions for each variable
4. **Monitoring** - Track performance metrics in production

### Testing
1. **Testable Architecture** - Use interfaces and dependency injection
2. **Mock Implementations** - Test without requiring actual SimConnect
3. **Integration Tests** - Test with real SimConnect when possible
4. **Error Scenarios** - Test failure modes and recovery

## Related Documentation

- [Performance Optimization](performance.md) - Detailed performance tuning guide
- [Troubleshooting Guide](troubleshooting.md) - Common issues and solutions
- [API Reference](../api/) - Complete API documentation
- [Examples](../examples/) - Practical implementation examples
