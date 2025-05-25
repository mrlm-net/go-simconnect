package client

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// FlightVariable represents a simulation variable definition
type FlightVariable struct {
	Name    string    // Human-readable name
	SimVar  string    // SimConnect variable name
	Units   string    // Units of measurement
	Value   float64   // Current value
	Updated time.Time // Last update time
}

// FlightDataManager manages real-time flight simulation data using separate data definitions
type FlightDataManager struct {
	client      *Client
	variables   []FlightVariable
	definitions []DataDefinitionID
	requests    []SimObjectDataRequestID
	mutex       sync.RWMutex
	running     bool
	stopChan    chan bool
	errorChan   chan error
	dataCount   int64
	errorCount  int64
	lastUpdate  time.Time
}

// NewFlightDataManager creates a new flight data manager
func NewFlightDataManager(client *Client) *FlightDataManager {
	return &FlightDataManager{
		client:    client,
		stopChan:  make(chan bool),
		errorChan: make(chan error, 10), // Buffered channel for errors
	}
}

// AddVariable adds a simulation variable to be tracked
func (fdm *FlightDataManager) AddVariable(name, simVar, units string) error {
	fdm.mutex.Lock()
	defer fdm.mutex.Unlock()

	if fdm.running {
		return fmt.Errorf("cannot add variables while data manager is running")
	} // Create unique IDs for this variable with proper spacing to avoid conflicts
	// According to Microsoft docs, RequestID should be unique for each request
	// Using larger, more predictable gaps to ensure no collisions
	index := len(fdm.variables)
	defineID := DataDefinitionID(index + 1)
	// Use even larger gaps for RequestIDs - start at 1000 and increment by 1000
	// This ensures no conflicts and makes debugging easier
	requestID := SimObjectDataRequestID(1000 + (index * 1000))

	// Add to SimConnect data definition
	if err := fdm.client.AddToDataDefinition(defineID, simVar, units, SIMCONNECT_DATATYPE_FLOAT64); err != nil {
		return fmt.Errorf("failed to add variable %s: %v", name, err)
	}

	// Create variable record
	variable := FlightVariable{
		Name:   name,
		SimVar: simVar,
		Units:  units,
		Value:  0.0,
	} // Store in our collections
	fdm.variables = append(fdm.variables, variable)
	fdm.definitions = append(fdm.definitions, defineID)
	fdm.requests = append(fdm.requests, requestID)

	return nil
}

// AddStandardVariables adds a set of commonly used flight simulation variables
func (fdm *FlightDataManager) AddStandardVariables() error {
	standardVars := []struct {
		name   string
		simVar string
		units  string
	}{
		{"Altitude", "Plane Altitude", "feet"},
		{"Indicated Airspeed", "Airspeed Indicated", "knots"},
		{"True Airspeed", "Airspeed True", "knots"},
		{"Ground Speed", "Ground Velocity", "knots"},
		{"Latitude", "Plane Latitude", "degrees"},
		{"Longitude", "Plane Longitude", "degrees"},
		{"Heading Magnetic", "Plane Heading Degrees Magnetic", "degrees"},
		{"Heading True", "Plane Heading Degrees True", "degrees"},
		{"Bank Angle", "Plane Bank Degrees", "degrees"},
		{"Pitch Angle", "Plane Pitch Degrees", "degrees"},
		{"Vertical Speed", "Vertical Speed", "feet per minute"},
		{"Engine RPM", "General Eng RPM:1", "rpm"},
		{"Throttle Position", "General Eng Throttle Lever Position:1", "percent"},
		{"Gear Position", "Gear Handle Position", "bool"},
		{"Flaps Position", "Flaps Handle Percent", "percent"},
	}

	for _, variable := range standardVars {
		if err := fdm.AddVariable(variable.name, variable.simVar, variable.units); err != nil {
			return err
		}
	}

	return nil
}

// Start begins real-time data collection
func (fdm *FlightDataManager) Start() error {
	fdm.mutex.Lock()
	defer fdm.mutex.Unlock()

	if fdm.running {
		return fmt.Errorf("data manager is already running")
	}

	if len(fdm.variables) == 0 {
		return fmt.Errorf("no variables added")
	} // Request data for all variables using optimized settings based on Microsoft SimConnect documentation
	// Using SIMCONNECT_PERIOD_SECOND for consistent 1Hz updates and CHANGED flag to reduce unnecessary data transmission
	// This combination provides the best performance for flight data monitoring applications
	for i, requestID := range fdm.requests {
		if err := fdm.client.RequestDataOnSimObjectWithFlags(
			requestID,
			fdm.definitions[i],
			SIMCONNECT_OBJECT_ID_USER,
			SIMCONNECT_PERIOD_SECOND,
			SIMCONNECT_DATA_REQUEST_FLAG_CHANGED,
			0, // origin (unused for SECOND period)
			0, // interval (unused for SECOND period)
			0, // limit (unused for SECOND period)
		); err != nil {
			return fmt.Errorf("failed to request data for variable %s: %v", fdm.variables[i].Name, err)
		}
		fmt.Printf("DEBUG: Requested data for %s with RequestID %d, DefineID %d using PERIOD_SECOND + CHANGED flag\n",
			fdm.variables[i].Name, requestID, fdm.definitions[i])
	}

	fdm.running = true

	// Start background data collection
	go fdm.dataCollectionLoop()

	return nil
}

// Stop stops real-time data collection
func (fdm *FlightDataManager) Stop() {
	fdm.mutex.Lock()
	defer fdm.mutex.Unlock()

	if !fdm.running {
		return
	}

	fdm.running = false
	fdm.stopChan <- true
}

// GetVariable returns the current value of a variable by name
func (fdm *FlightDataManager) GetVariable(name string) (FlightVariable, bool) {
	fdm.mutex.RLock()
	defer fdm.mutex.RUnlock()

	for _, variable := range fdm.variables {
		if variable.Name == name {
			return variable, true
		}
	}

	return FlightVariable{}, false
}

// GetAllVariables returns all current variable values
func (fdm *FlightDataManager) GetAllVariables() []FlightVariable {
	fdm.mutex.RLock()
	defer fdm.mutex.RUnlock()

	// Return current values from the variables array
	result := make([]FlightVariable, len(fdm.variables))
	copy(result, fdm.variables)
	return result
}

// GetStats returns data collection statistics
func (fdm *FlightDataManager) GetStats() (dataCount int64, errorCount int64, lastUpdate time.Time) {
	fdm.mutex.RLock()
	defer fdm.mutex.RUnlock()
	return fdm.dataCount, fdm.errorCount, fdm.lastUpdate
}

// GetErrors returns a channel for receiving errors (non-blocking)
func (fdm *FlightDataManager) GetErrors() <-chan error {
	return fdm.errorChan
}

// IsRunning returns whether the data manager is currently collecting data
func (fdm *FlightDataManager) IsRunning() bool {
	fdm.mutex.RLock()
	defer fdm.mutex.RUnlock()
	return fdm.running
}

// dataCollectionLoop runs in a separate goroutine to collect data
func (fdm *FlightDataManager) dataCollectionLoop() {
	for {
		select {
		case <-fdm.stopChan:
			return
		default:
			fdm.collectData()
			time.Sleep(50 * time.Millisecond) // Limit frequency
		}
	}
}

// collectData collects a single round of data from SimConnect
func (fdm *FlightDataManager) collectData() {
	data, err := fdm.client.GetRawDispatch()
	if err != nil {
		return
	}
	if data == nil {
		return
	}

	msgType, err := ParseMessageType(data)
	if err != nil {
		fdm.errorCount++
		select {
		case fdm.errorChan <- err:
		default: // Channel full, drop error
		}
		return
	}

	if msgType == SIMCONNECT_RECV_ID_SIMOBJECT_DATA {
		header, simData, err := ParseSimObjectData(data)
		if err != nil {
			fdm.errorCount++
			select {
			case fdm.errorChan <- err:
			default: // Channel full, drop error
			}
			return
		}

		if len(simData) < 8 {
			return
		}
		// Find the variable this data corresponds to
		requestID := SimObjectDataRequestID(header.DwRequestID)
		fdm.mutex.Lock()

		// Find the variable index by requestID
		variableIndex := -1
		for i, reqID := range fdm.requests {
			if reqID == requestID {
				variableIndex = i
				break
			}
		}

		// Debug logging to track what's happening
		if variableIndex >= 0 {
			// Successfully found variable
		} else {
			// This indicates a RequestID mismatch - log it for debugging
			fmt.Printf("DEBUG: Received unknown RequestID %d, known RequestIDs: %v\n", requestID, fdm.requests)
		}
		if variableIndex >= 0 && variableIndex < len(fdm.variables) {
			value := *(*float64)(unsafe.Pointer(&simData[0]))
			// Update the variable directly in the slice
			fdm.variables[variableIndex].Value = value
			fdm.variables[variableIndex].Updated = time.Now()
			fdm.dataCount++
			fdm.lastUpdate = time.Now()

			// Debug: Log which variable was updated
			fmt.Printf("DEBUG: Updated %s = %.2f (RequestID: %d)\n",
				fdm.variables[variableIndex].Name, value, requestID)
		}
		fdm.mutex.Unlock()
	}
}
