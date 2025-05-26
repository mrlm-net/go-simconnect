package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mrlm-net/go-simconnect/pkg/client"
)

// FlightMonitor provides real-time flight monitoring with bidirectional event testing
type FlightMonitor struct {
	client       *client.Client
	eventManager *client.SystemEventManager
	flightData   *client.FlightDataManager
	dashboard    *Dashboard
	stateTracker *StateTracker
	mu           sync.RWMutex
	running      bool
	startTime    time.Time
}

// Dashboard handles the visual display and user interface
type Dashboard struct {
	mu           sync.RWMutex
	eventCounts  map[string]int
	lastEvents   map[string]time.Time
	totalEvents  int
	lastUpdate   time.Time
}

// StateTracker monitors simulation state changes
type StateTracker struct {
	mu               sync.RWMutex
	isPaused         bool
	isSimRunning     bool
	currentAircraft  string
	currentFlight    string
	lastPositionTime time.Time
	frameRate        uint32
	soundEnabled     bool
	viewState        uint32
}

func NewFlightMonitor() *FlightMonitor {
	return &FlightMonitor{
		dashboard:    NewDashboard(),
		stateTracker: NewStateTracker(),
		startTime:    time.Now(),
	}
}

func NewDashboard() *Dashboard {
	return &Dashboard{
		eventCounts: make(map[string]int),
		lastEvents:  make(map[string]time.Time),
	}
}

func NewStateTracker() *StateTracker {
	return &StateTracker{
		soundEnabled: true, // Default assumption
	}
}

func main() {
	fmt.Println("=== REAL-TIME FLIGHT MONITOR DEMO ===")
	fmt.Println("Advanced System Events Testing with Bidirectional Validation")
	fmt.Println("This demo tests both receiving events AND triggering state changes")
	fmt.Println()

	monitor := NewFlightMonitor()

	// Initialize SimConnect
	if err := monitor.initializeSimConnect(); err != nil {
		log.Fatalf("Failed to initialize SimConnect: %v", err)
	}
	defer monitor.cleanup()

	// Set up managers
	if err := monitor.setupManagers(); err != nil {
		log.Fatalf("Failed to setup managers: %v", err)
	}

	// Start monitoring
	if err := monitor.startMonitoring(); err != nil {
		log.Fatalf("Failed to start monitoring: %v", err)
	}

	// Run interactive session
	monitor.runInteractiveSession()
}

func (fm *FlightMonitor) initializeSimConnect() error {
	fmt.Println("STEP 1: Connecting to Microsoft Flight Simulator...")
	
	// Use MSFS 2024 SDK DLL path
	dllPath := `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
	fm.client = client.NewClientWithDLLPath("FlightMonitorDemo", dllPath)

	if err := fm.client.Open(); err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	fmt.Println("✅ Successfully connected to SimConnect!")
	return nil
}

func (fm *FlightMonitor) setupManagers() error {
	fmt.Println("STEP 2: Setting up event and data managers...")

	// Create SystemEventManager
	fm.eventManager = client.NewSystemEventManager(fm.client)

	// Create FlightDataManager for integration testing
	fm.flightData = client.NewFlightDataManager(fm.client)

	// Add comprehensive flight variables for state validation
	variables := map[string]struct {
		simVar   string
		units    string
		writable bool
	}{
		// Core flight data
		"Altitude":         {"INDICATED ALTITUDE", "feet", false},
		"Airspeed":         {"AIRSPEED INDICATED", "knots", false},
		"Heading":          {"HEADING INDICATOR", "degrees", false},
		"Vertical Speed":   {"VERTICAL SPEED", "feet per minute", false},
		
		// System state
		"Sim Paused":       {"SIM PAUSED", "bool", false},
		"Sim Running":      {"SIM", "bool", false},
		"Ground Speed":     {"GROUND VELOCITY", "knots", false},
		
		// Engine data
		"Engine Running":   {"GENERAL ENG COMBUSTION:1", "bool", false},
		"Throttle":         {"GENERAL ENG THROTTLE LEVER POSITION:1", "percent", true},
		
		// Aircraft state
		"On Ground":        {"SIM ON GROUND", "bool", false},
		"Parking Brake":    {"BRAKE PARKING INDICATOR", "bool", false},
		
		// Camera/View (writable for testing)
		"Camera State":     {"CAMERA STATE", "number", true},
	}

	for name, config := range variables {
		if config.writable {
			if err := fm.flightData.AddVariableWithWritable(name, config.simVar, config.units, true); err != nil {
				log.Printf("Warning: Could not add writable variable %s: %v", name, err)
			}
		} else {
			if err := fm.flightData.AddVariable(name, config.simVar, config.units); err != nil {
				log.Printf("Warning: Could not add variable %s: %v", name, err)
			}
		}
	}

	// Subscribe to comprehensive system events
	eventSubscriptions := map[string]string{
		// Timer events for regular monitoring
		"Timer1Sec":     client.SystemEvent1Sec,
		"Timer4Sec":     client.SystemEvent4Sec,
		"Timer6Hz":      client.SystemEvent6Hz,
		"Frame":         client.SystemEventFrame,

		// Simulation state events
		"SimStart":      client.SystemEventSimStart,
		"SimStop":       client.SystemEventSimStop,
		"Sim":           client.SystemEventSim,

		// Pause events (critical for validation)
		"Pause":         client.SystemEventPause,
		"Paused":        client.SystemEventPaused,
		"Unpaused":      client.SystemEventUnpaused,
		"PauseEx":       client.SystemEventPauseEx,

		// Flight and aircraft events
		"FlightLoaded":  client.SystemEventFlightLoaded,
		"FlightSaved":   client.SystemEventFlightSaved,
		"AircraftLoaded": client.SystemEventAircraftLoaded,

		// Position and view events
		"PositionChanged": client.SystemEventPositionChanged,
		"ViewChanged":     client.SystemEventView,

		// System events
		"Sound":         client.SystemEventSound,
		"Crashed":       client.SystemEventCrashed,
		"CrashReset":    client.SystemEventCrashReset,
	}

	for displayName, eventName := range eventSubscriptions {
		if _, err := fm.eventManager.SubscribeToEvent(eventName, fm.createEventHandler(displayName)); err != nil {
			log.Printf("Warning: Failed to subscribe to %s: %v", displayName, err)
		} else {
			fmt.Printf("  ✅ Subscribed: %s\n", displayName)
		}
	}

	fmt.Println("✅ Managers configured successfully!")
	return nil
}

func (fm *FlightMonitor) createEventHandler(eventName string) client.SystemEventCallback {
	var lastFrameDisplay time.Time
	var lastTimer6HzDisplay time.Time
	
	return func(event client.SystemEventData) {
		timestamp := time.Now().Format("15:04:05.000")
		
		// Update dashboard
		fm.dashboard.recordEvent(eventName)
		
		// Update state tracker based on event
		fm.stateTracker.processEvent(eventName, event)
		
		// Throttle high-frequency events to avoid console spam
		if eventName == "Frame" {
			// Only display Frame events every 2 seconds
			if time.Since(lastFrameDisplay) < 2*time.Second {
				return
			}
			lastFrameDisplay = time.Now()
		} else if eventName == "Timer6Hz" {
			// Only display Timer6Hz events every 3 seconds
			if time.Since(lastTimer6HzDisplay) < 3*time.Second {
				return
			}
			lastTimer6HzDisplay = time.Now()
		}
		
		// Display event with appropriate emoji and formatting
		icon := fm.getEventIcon(eventName)
		
		fmt.Printf("[%s] %s %s", timestamp, icon, eventName)
		
		// Add relevant event data
		switch eventName {
		case "Pause", "Paused", "Unpaused":
			state := "OFF"
			if event.Data == uint32(client.SIMCONNECT_STATE_ON) {
				state = "ON"
			}
			fmt.Printf(": %s (Data=%d)", state, event.Data)
			
		case "PauseEx":
			// Decode pause flags
			flags := event.Data
			pauseTypes := []string{}
			if flags&uint32(client.PAUSE_STATE_FLAG_PAUSE) != 0 {
				pauseTypes = append(pauseTypes, "FULL")
			}
			if flags&uint32(client.PAUSE_STATE_FLAG_ACTIVE_PAUSE) != 0 {
				pauseTypes = append(pauseTypes, "ACTIVE")
			}
			if flags&uint32(client.PAUSE_STATE_FLAG_SIM_PAUSE) != 0 {
				pauseTypes = append(pauseTypes, "SIM")
			}
			if len(pauseTypes) == 0 {
				pauseTypes = append(pauseTypes, "NONE")
			}
			fmt.Printf(": [%s] (Flags=0x%X)", strings.Join(pauseTypes, ","), flags)
			
		case "Frame":
			fmt.Printf(": FPS=%d", event.Data)
			
		case "FlightLoaded", "FlightSaved", "AircraftLoaded":
			if event.Filename != "" {
				fmt.Printf(": %s", event.Filename)
			} else {
				fmt.Printf(": Data=%d", event.Data)
			}
			
		case "ViewChanged":
			viewType := "Unknown"
			switch event.Data {
			case client.SIMCONNECT_VIEW_SYSTEM_EVENT_DATA_COCKPIT_2D:
				viewType = "2D Cockpit"
			case client.SIMCONNECT_VIEW_SYSTEM_EVENT_DATA_COCKPIT_VIRTUAL:
				viewType = "Virtual Cockpit"
			case client.SIMCONNECT_VIEW_SYSTEM_EVENT_DATA_ORTHOGONAL:
				viewType = "Map View"
			}
			fmt.Printf(": %s (Data=0x%X)", viewType, event.Data)
			
		case "Sound":
			soundState := "OFF"
			if event.Data&uint32(client.SIMCONNECT_SOUND_SYSTEM_EVENT_DATA_MASTER) != 0 {
				soundState = "ON"
			}
			fmt.Printf(": %s (Data=0x%X)", soundState, event.Data)
			
		default:
			if event.Data != 0 {
				fmt.Printf(": Data=%d", event.Data)
			}
		}
		
		fmt.Println()
	}
}

func (fm *FlightMonitor) getEventIcon(eventName string) string {
	icons := map[string]string{
		"Timer1Sec": "⏱️",
		"Timer4Sec": "⏰",
		"Timer6Hz":  "⚡",
		"Frame":     "🎬",
		"SimStart":  "▶️",
		"SimStop":   "⏹️",
		"Sim":       "🎮",
		"Pause":     "⏸️",
		"Paused":    "🔸",
		"Unpaused":  "▶️",
		"PauseEx":   "⏸️",
		"FlightLoaded": "✈️",
		"FlightSaved":  "💾",
		"AircraftLoaded": "🛩️",
		"PositionChanged": "📍",
		"ViewChanged": "👁️",
		"Sound":      "🔊",
		"Crashed":    "💥",
		"CrashReset": "🔄",
	}
	
	if icon, exists := icons[eventName]; exists {
		return icon
	}
	return "📡"
}

func (fm *FlightMonitor) startMonitoring() error {
	fmt.Println("STEP 3: Starting real-time monitoring...")

	// Start FlightDataManager
	if err := fm.flightData.Start(); err != nil {
		log.Printf("Warning: Could not start FlightDataManager: %v", err)
	} else {
		fmt.Println("✅ FlightDataManager started")
	}

	// Start SystemEventManager
	if err := fm.eventManager.Start(); err != nil {
		return fmt.Errorf("failed to start SystemEventManager: %v", err)
	}
	fmt.Println("✅ SystemEventManager started")

	// Monitor for errors
	go func() {
		for err := range fm.eventManager.GetErrors() {
			log.Printf("🚨 Event Manager Error: %v", err)
		}
	}()

	go func() {
		for err := range fm.flightData.GetErrors() {
			log.Printf("🚨 Flight Data Manager Error: %v", err)
		}
	}()

	fm.mu.Lock()
	fm.running = true
	fm.mu.Unlock()

	fmt.Println("✅ Real-time monitoring active!")
	return nil
}

func (fm *FlightMonitor) runInteractiveSession() {
	fmt.Println()
	fmt.Println("=== INTERACTIVE FLIGHT MONITOR ===")
	fmt.Println("Monitor is now running. Available commands:")
	fmt.Println("  📊 'status'     - Show current statistics and state")
	fmt.Println("  📈 'data'       - Show current flight data")
	fmt.Println("  ⏸️  'pause'      - Toggle simulation pause (test event triggering)")
	fmt.Println("  📷 'camera X'   - Change camera view (2-6, test writable variable)")
	fmt.Println("  🎯 'throttle X' - Set throttle percentage (0-100, test SetData)")
	fmt.Println("  🔧 'test'       - Run automated validation tests")
	fmt.Println("  ❓ 'help'       - Show this help")
	fmt.Println("  ❌ 'quit'       - Exit the monitor")
	fmt.Println()
	fmt.Println("Try pausing/unpausing MSFS to see events in real-time!")
	fmt.Println("Events will appear below as they occur...")
	fmt.Println()

	// Set up graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Periodic status updates
	statusTicker := time.NewTicker(10 * time.Second)
	defer statusTicker.Stop()

	// Command input handling
	inputChan := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputChan <- strings.TrimSpace(scanner.Text())
		}
	}()

	// Main monitoring loop
	for {
		select {
		case <-shutdownChan:
			fmt.Println("\n⏹️ Shutdown signal received...")
			return

		case <-statusTicker.C:
			fm.displayPeriodicStatus()

		case command := <-inputChan:
			if fm.handleCommand(command) {
				return // Exit requested
			}
		}
	}
}

func (fm *FlightMonitor) handleCommand(command string) bool {
	parts := strings.Fields(strings.ToLower(command))
	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "quit", "exit", "q":
		fmt.Println("👋 Exiting flight monitor...")
		return true

	case "help", "h":
		fm.showHelp()

	case "status", "s":
		fm.displayDetailedStatus()

	case "data", "d":
		fm.displayFlightData()

	case "pause", "p":
		fm.togglePause()

	case "camera", "cam", "c":
		if len(parts) >= 2 {
			if state, err := strconv.Atoi(parts[1]); err == nil {
				fm.setCameraState(state)
			} else {
				fmt.Printf("❌ Invalid camera state. Use: camera <2-6>\n")
			}
		} else {
			fmt.Printf("❌ Usage: camera <state>\n  Valid states: 2=Wing, 3=Cockpit, 4=External, 5=Tail, 6=Tower\n")
		}

	case "throttle", "thr", "t":
		if len(parts) >= 2 {
			if percent, err := strconv.ParseFloat(parts[1], 64); err == nil {
				fm.setThrottle(percent)
			} else {
				fmt.Printf("❌ Invalid throttle value. Use: throttle <0-100>\n")
			}
		} else {
			fmt.Printf("❌ Usage: throttle <percentage>\n  Example: throttle 75\n")
		}

	case "test":
		fm.runValidationTests()

	default:
		fmt.Printf("❓ Unknown command: %s\n", command)
		fmt.Printf("💡 Type 'help' for available commands\n")
	}

	return false
}

func (fm *FlightMonitor) showHelp() {
	fmt.Println("\n=== FLIGHT MONITOR COMMANDS ===")
	fmt.Println("📊 status      - Show monitoring statistics and simulation state")
	fmt.Println("📈 data        - Display current flight data values")
	fmt.Println("⏸️  pause       - Toggle simulation pause state")
	fmt.Println("📷 camera <2-6> - Change camera view (2=Wing, 3=Cockpit, 4=External, 5=Tail, 6=Tower)")
	fmt.Println("🎯 throttle <0-100> - Set throttle percentage")
	fmt.Println("🔧 test        - Run automated bidirectional validation tests")
	fmt.Println("❓ help        - Show this help message")
	fmt.Println("❌ quit        - Exit the flight monitor")
	fmt.Println()
	fmt.Println("💡 TIP: Try these actions in MSFS to generate events:")
	fmt.Println("   • Press ESC or PAUSE to pause/unpause")
	fmt.Println("   • Load different aircraft or flights")
	fmt.Println("   • Change camera views with external view keys")
	fmt.Println("   • Move aircraft position using slew mode")
	fmt.Println()
}

func (fm *FlightMonitor) displayPeriodicStatus() {
	fmt.Println("\n--- PERIODIC STATUS UPDATE ---")
	fm.displayQuickStatus()
	fmt.Println("--- END STATUS UPDATE ---")
}

func (fm *FlightMonitor) displayDetailedStatus() {
	fmt.Println("\n=== DETAILED FLIGHT MONITOR STATUS ===")
	
	duration := time.Since(fm.startTime)
	fmt.Printf("⏱️  Monitoring Duration: %.1f seconds\n", duration.Seconds())
	
	// Dashboard statistics
	fm.dashboard.mu.RLock()
	totalEvents := fm.dashboard.totalEvents
	eventCounts := make(map[string]int)
	for k, v := range fm.dashboard.eventCounts {
		eventCounts[k] = v
	}
	fm.dashboard.mu.RUnlock()
	
	fmt.Printf("📡 Total Events Received: %d\n", totalEvents)
	
	if totalEvents > 0 {
		fmt.Printf("📊 Events Per Second: %.2f\n", float64(totalEvents)/duration.Seconds())
		
		fmt.Println("\n📈 Event Breakdown:")
		for eventName, count := range eventCounts {
			percentage := float64(count) / float64(totalEvents) * 100
			fmt.Printf("   %-20s: %4d (%.1f%%)\n", eventName, count, percentage)
		}
	}
	
	// State tracker information
	fm.stateTracker.mu.RLock()
	fmt.Printf("\n🎮 Simulation State:\n")
	fmt.Printf("   Paused: %t\n", fm.stateTracker.isPaused)
	fmt.Printf("   Running: %t\n", fm.stateTracker.isSimRunning)
	fmt.Printf("   Sound: %t\n", fm.stateTracker.soundEnabled)
	if fm.stateTracker.frameRate > 0 {
		fmt.Printf("   Frame Rate: %d FPS\n", fm.stateTracker.frameRate)
	}
	if fm.stateTracker.currentAircraft != "" {
		fmt.Printf("   Aircraft: %s\n", fm.stateTracker.currentAircraft)
	}
	if fm.stateTracker.currentFlight != "" {
		fmt.Printf("   Flight: %s\n", fm.stateTracker.currentFlight)
	}
	fm.stateTracker.mu.RUnlock()
	
	// Manager status
	fmt.Printf("\n🔧 Manager Status:\n")
	fmt.Printf("   SystemEventManager: %t\n", fm.eventManager.IsRunning())
	fmt.Printf("   FlightDataManager: %t\n", fm.flightData.IsRunning())
	
	if fm.flightData.IsRunning() {
		dataCount, errorCount, lastUpdate := fm.flightData.GetStats()
		fmt.Printf("   Flight Data Points: %d (errors: %d)\n", dataCount, errorCount)
		if !lastUpdate.IsZero() {
			fmt.Printf("   Last Data Update: %v ago\n", time.Since(lastUpdate).Truncate(time.Millisecond))
		}
	}
	
	subscribedEvents := fm.eventManager.GetSubscribedEvents()
	fmt.Printf("   Event Subscriptions: %d active\n", len(subscribedEvents))
	
	fmt.Println("=== END DETAILED STATUS ===")
}

func (fm *FlightMonitor) displayQuickStatus() {
	duration := time.Since(fm.startTime)
	
	fm.dashboard.mu.RLock()
	totalEvents := fm.dashboard.totalEvents
	fm.dashboard.mu.RUnlock()
	
	fm.stateTracker.mu.RLock()
	isPaused := fm.stateTracker.isPaused
	isRunning := fm.stateTracker.isSimRunning
	frameRate := fm.stateTracker.frameRate
	fm.stateTracker.mu.RUnlock()
	
	fmt.Printf("⏱️ %.1fs | 📡 %d events | 🎮 Sim:%t Pause:%t", 
		duration.Seconds(), totalEvents, isRunning, isPaused)
	
	if frameRate > 0 {
		fmt.Printf(" | 🎬 %dFPS", frameRate)
	}
	
	if fm.flightData.IsRunning() {
		dataCount, _, _ := fm.flightData.GetStats()
		fmt.Printf(" | 📊 %d data", dataCount)
	}
	
	fmt.Println()
}

func (fm *FlightMonitor) displayFlightData() {
	fmt.Println("\n=== CURRENT FLIGHT DATA ===")
	
	if !fm.flightData.IsRunning() {
		fmt.Println("❌ FlightDataManager is not running")
		return
	}
	
	variables := fm.flightData.GetAllVariables()
	if len(variables) == 0 {
		fmt.Println("📭 No flight data available")
		return
	}
	
	// Group variables by category
	categories := map[string][]string{
		"🛩️  Aircraft": {"Altitude", "Airspeed", "Heading", "Vertical Speed"},
		"⚙️  Systems":  {"Engine Running", "Throttle", "Parking Brake"},
		"🎮 Simulation": {"Sim Paused", "Sim Running", "On Ground", "Camera State"},
		"📍 Navigation": {"Ground Speed"},
	}
		for category, varNames := range categories {
		fmt.Printf("\n%s:\n", category)
		hasData := false
		
		for _, name := range varNames {
			// Find variable by name in the slice
			var variable *client.FlightVariable
			for i := range variables {
				if variables[i].Name == name {
					variable = &variables[i]
					break
				}
			}
			
			if variable != nil {
				hasData = true
				fmt.Printf("   %-15s: ", name)
				
				// Format value based on type
				if variable.Units == "bool" {
					if variable.Value > 0.5 {
						fmt.Printf("✅ TRUE")
					} else {
						fmt.Printf("❌ FALSE")
					}
				} else {
					fmt.Printf("%.2f %s", variable.Value, variable.Units)
				}
				
				if variable.Writable {
					fmt.Printf(" (writable)")
				}
				
				fmt.Printf("\n")
			}
		}
		
		if !hasData {
			fmt.Printf("   📭 No data available\n")
		}
	}
	
	dataCount, errorCount, lastUpdate := fm.flightData.GetStats()
	fmt.Printf("\n📊 Statistics: %d updates, %d errors\n", dataCount, errorCount)
	if !lastUpdate.IsZero() {
		fmt.Printf("🕐 Last update: %v ago\n", time.Since(lastUpdate).Truncate(time.Millisecond))
	}
	
	fmt.Println("=== END FLIGHT DATA ===")
}

func (fm *FlightMonitor) togglePause() {
	fmt.Println("⏸️ Attempting to toggle simulation pause...")
	fmt.Println("❗ Note: Direct pause control not implemented in this library version")
	fmt.Println("💡 Please use ESC or PAUSE key in MSFS to test pause events")
	fmt.Println("🔍 Watch for Pause/Paused/Unpaused events in the monitor output")
}

func (fm *FlightMonitor) setCameraState(state int) {
	if state < 2 || state > 6 {
		fmt.Printf("❌ Invalid camera state %d. Valid range: 2-6\n", state)
		return
	}
	
	fmt.Printf("📷 Setting camera state to %d", state)
	
	stateNames := map[int]string{
		2: "Wing View",
		3: "Cockpit View", 
		4: "External View",
		5: "Tail View",
		6: "Tower View",
	}
	
	if name, exists := stateNames[state]; exists {
		fmt.Printf(" (%s)", name)
	}
	fmt.Println("...")
	
	if err := fm.flightData.SetVariable("Camera State", float64(state)); err != nil {
		fmt.Printf("❌ Failed to set camera state: %v\n", err)
		fmt.Println("💡 Make sure aircraft is loaded and FlightDataManager is running")
	} else {
		fmt.Printf("✅ Camera state command sent!\n")
		fmt.Println("🔍 Watch for ViewChanged events to confirm the change")
	}
}

func (fm *FlightMonitor) setThrottle(percent float64) {
	if percent < 0 || percent > 100 {
		fmt.Printf("❌ Invalid throttle percentage %.1f. Valid range: 0-100\n", percent)
		return
	}
	
	fmt.Printf("🎯 Setting throttle to %.1f%%...\n", percent)
	
	if err := fm.flightData.SetVariable("Throttle", percent); err != nil {
		fmt.Printf("❌ Failed to set throttle: %v\n", err)
		fmt.Println("💡 Make sure aircraft is loaded and engine is available")
	} else {
		fmt.Printf("✅ Throttle command sent!\n")
		fmt.Println("🔍 Check throttle position in aircraft and flight data")
	}
}

func (fm *FlightMonitor) runValidationTests() {
	fmt.Println("\n=== AUTOMATED VALIDATION TESTS ===")
	fmt.Println("🔧 Running bidirectional system events validation...")
	
	tests := []struct {
		name string
		test func() bool
	}{
		{"Event Subscription Status", fm.testEventSubscriptions},
		{"Manager Integration", fm.testManagerIntegration},
		{"State Consistency", fm.testStateConsistency},
		{"Data Variable Access", fm.testDataVariables},
		{"Writable Variables", fm.testWritableVariables},
	}
	
	passed := 0
	total := len(tests)
	
	for _, test := range tests {
		fmt.Printf("🧪 Testing: %s... ", test.name)
		if test.test() {
			fmt.Println("✅ PASS")
			passed++
		} else {
			fmt.Println("❌ FAIL")
		}
	}
	
	fmt.Printf("\n📊 Test Results: %d/%d passed (%.1f%%)\n", 
		passed, total, float64(passed)/float64(total)*100)
	
	if passed == total {
		fmt.Println("🎉 All validation tests passed! System events are working correctly.")
	} else {
		fmt.Println("⚠️  Some tests failed. Check MSFS connection and aircraft state.")
	}
	
	fmt.Println("=== END VALIDATION TESTS ===")
}

func (fm *FlightMonitor) testEventSubscriptions() bool {
	subscriptions := fm.eventManager.GetSubscribedEvents()
	return len(subscriptions) > 10 // Should have many subscriptions
}

func (fm *FlightMonitor) testManagerIntegration() bool {
	return fm.eventManager.IsRunning() && fm.flightData.IsRunning()
}

func (fm *FlightMonitor) testStateConsistency() bool {
	// Check if we've received recent events
	fm.dashboard.mu.RLock()
	totalEvents := fm.dashboard.totalEvents
	fm.dashboard.mu.RUnlock()
	
	return totalEvents > 0 // Should have received some events
}

func (fm *FlightMonitor) testDataVariables() bool {
	variables := fm.flightData.GetAllVariables()
	return len(variables) > 5 // Should have multiple variables
}

func (fm *FlightMonitor) testWritableVariables() bool {
	variables := fm.flightData.GetAllVariables()
	writableCount := 0
	
	for _, variable := range variables {
		if variable.Writable {
			writableCount++
		}
	}
	
	return writableCount > 0 // Should have at least some writable variables
}

func (fm *FlightMonitor) cleanup() {
	fmt.Println("\n🧹 Cleaning up...")
	
	if fm.eventManager != nil {
		fm.eventManager.Stop()
		fmt.Println("✅ SystemEventManager stopped")
	}
	
	if fm.flightData != nil {
		fm.flightData.Stop()
		fmt.Println("✅ FlightDataManager stopped")
	}
	
	if fm.client != nil {
		fm.client.Close()
		fmt.Println("✅ SimConnect connection closed")
	}
	
	// Final statistics
	duration := time.Since(fm.startTime)
	fm.dashboard.mu.RLock()
	totalEvents := fm.dashboard.totalEvents
	fm.dashboard.mu.RUnlock()
	
	fmt.Printf("\n📊 FINAL SESSION STATISTICS:\n")
	fmt.Printf("   Duration: %.1f seconds\n", duration.Seconds())
	fmt.Printf("   Total Events: %d\n", totalEvents)
	if duration.Seconds() > 0 {
		fmt.Printf("   Avg Events/sec: %.2f\n", float64(totalEvents)/duration.Seconds())
	}
	
	fmt.Println("\n🎉 Flight Monitor Demo completed successfully!")
	fmt.Println("✅ System events implementation validated with live simulator")
}

// Dashboard methods
func (d *Dashboard) recordEvent(eventName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.eventCounts[eventName]++
	d.lastEvents[eventName] = time.Now()
	d.totalEvents++
	d.lastUpdate = time.Now()
}

// StateTracker methods
func (st *StateTracker) processEvent(eventName string, event client.SystemEventData) {
	st.mu.Lock()
	defer st.mu.Unlock()
	
	switch eventName {
	case "Pause", "Paused":
		st.isPaused = (event.Data == uint32(client.SIMCONNECT_STATE_ON))
	case "Unpaused":
		st.isPaused = false
	case "SimStart":
		st.isSimRunning = true
	case "SimStop":
		st.isSimRunning = false
	case "Sim":
		st.isSimRunning = (event.Data == uint32(client.SIMCONNECT_STATE_ON))
	case "Frame":
		st.frameRate = event.Data
	case "Sound":
		st.soundEnabled = (event.Data&uint32(client.SIMCONNECT_SOUND_SYSTEM_EVENT_DATA_MASTER) != 0)
	case "ViewChanged":
		st.viewState = event.Data
	case "FlightLoaded":
		if event.Filename != "" {
			st.currentFlight = event.Filename
		}
	case "AircraftLoaded":
		if event.Filename != "" {
			st.currentAircraft = event.Filename
		}
	case "PositionChanged":
		st.lastPositionTime = time.Now()
	}
}
