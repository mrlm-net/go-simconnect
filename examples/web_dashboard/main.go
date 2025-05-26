package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mrlm-net/go-simconnect/pkg/client"
)

type FlightData struct {
	// Connection status
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`

	// Flight variables
	Altitude        float64 `json:"altitude"`
	IndicatedSpeed  float64 `json:"indicatedSpeed"`
	GroundSpeed     float64 `json:"groundSpeed"`
	VerticalSpeed   float64 `json:"verticalSpeed"`
	HeadingMagnetic float64 `json:"headingMagnetic"`
	BankAngle       float64 `json:"bankAngle"`
	PitchAngle      float64 `json:"pitchAngle"`
	EngineRPM       float64 `json:"engineRPM"`
	ThrottlePos     float64 `json:"throttlePos"`
	GearPosition    float64 `json:"gearPosition"`
	FlapsPosition   float64 `json:"flapsPosition"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`

	// Weather variables
	AmbientTemperature float64 `json:"ambientTemperature"`
	BarometricPressure float64 `json:"barometricPressure"`
	WindSpeed          float64 `json:"windSpeed"`
	WindDirection      float64 `json:"windDirection"`
	Visibility         float64 `json:"visibility"`
	CloudCoverage      float64 `json:"cloudCoverage"`
	// Game information
	AircraftTitle  string  `json:"aircraftTitle"`
	SimulationRate float64 `json:"simulationRate"`
	LocalTime      string  `json:"localTime"`
	ZuluTime       string  `json:"zuluTime"`
	OnGround       bool    `json:"onGround"`
	ParkingBrake   bool    `json:"parkingBrake"`
	SimPaused      bool    `json:"simPaused"`

	// Statistics
	DataCount  int64     `json:"dataCount"`
	ErrorCount int64     `json:"errorCount"`
	LastUpdate time.Time `json:"lastUpdate"`
	UpdateRate float64   `json:"updateRate"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	}

	simclient *client.Client
	fdm       *client.FlightDataManager
	startTime time.Time
)

func main() {
	// Initialize SimConnect
	initSimConnect()
	// Setup HTTP routes
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("🚀 Starting MSFS 2024 Web Dashboard...")
	fmt.Println("📊 Dashboard: http://localhost:8080")
	fmt.Println("🔌 WebSocket: ws://localhost:8080/ws")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initSimConnect() {
	startTime = time.Now()

	// Create client with specific DLL path
	simclient = client.NewClientWithDLLPath("MSFS Web Dashboard", "C:\\MSFS 2024 SDK\\SimConnect SDK\\lib\\SimConnect.dll")

	// Try to connect
	if err := simclient.Open(); err != nil {
		log.Printf("⚠️  SimConnect connection failed: %v", err)
		log.Println("💡 Make sure MSFS 2024 is running and SimConnect is enabled")
		return
	}
	// Create flight data manager
	fdm = client.NewFlightDataManager(simclient)

	// Add flight data variables
	flightVars := []struct {
		name     string
		simVar   string
		units    string
		writable bool
	}{
		{"Altitude", "Plane Altitude", "feet", false},
		{"Indicated Airspeed", "Airspeed Indicated", "knots", false},
		{"True Airspeed", "Airspeed True", "knots", false},
		{"Ground Speed", "Ground Velocity", "knots", false},
		{"Latitude", "Plane Latitude", "degrees", false},
		{"Longitude", "Plane Longitude", "degrees", false},
		{"Heading Magnetic", "Plane Heading Degrees Magnetic", "degrees", false},
		{"Heading True", "Plane Heading Degrees True", "degrees", false},
		{"Bank Angle", "Plane Bank Degrees", "degrees", false},
		{"Pitch Angle", "Plane Pitch Degrees", "degrees", false},
		{"Vertical Speed", "Vertical Speed", "feet per minute", false},
		{"Engine RPM", "General Eng RPM:1", "rpm", false},
		{"Throttle Position", "General Eng Throttle Lever Position:1", "percent", true},
		{"Gear Position", "Gear Handle Position", "bool", true},
		{"Flaps Position", "Flaps Handle Percent", "percent", true},
	}

	for _, flightVar := range flightVars {
		if err := fdm.AddVariableWithWritable(flightVar.name, flightVar.simVar, flightVar.units, flightVar.writable); err != nil {
			log.Printf("⚠️  Warning: Failed to add flight variable %s: %v", flightVar.name, err)
		} else {
			log.Printf("✅ Successfully added flight variable: %s", flightVar.name)
		}
	}
	// Add weather variables
	weatherVars := []struct {
		name   string
		simVar string
		units  string
	}{
		{"Ambient Temperature", "Ambient Temperature", "celsius"},
		{"Barometric Pressure", "Kohlsman Setting HG", "inHg"},
		{"Wind Speed", "Ambient Wind Velocity", "knots"},
		{"Wind Direction", "Ambient Wind Direction", "degrees"},
		{"Visibility", "Ambient Visibility", "meters"},
		{"Cloud Coverage", "Ambient Total Cloud Coverage", "percent"},
	}
	for _, weather := range weatherVars {
		if err := fdm.AddVariable(weather.name, weather.simVar, weather.units); err != nil {
			log.Printf("⚠️  Warning: Failed to add weather variable %s: %v", weather.name, err)
		} else {
			log.Printf("✅ Successfully added weather variable: %s", weather.name)
		}
	} // Add game information variables
	gameVars := []struct {
		name   string
		simVar string
		units  string
	}{
		{"Aircraft Title", "ATC Type", "string"},
		{"Simulation Rate", "Simulation Rate", "number"},
		{"Local Time", "Local Time", "seconds"},
		{"Zulu Time", "Zulu Time", "seconds"},
		{"On Ground", "Sim On Ground", "bool"},
		{"Parking Brake", "Brake Parking Position", "bool"},
		{"Sim Paused", "Sim Paused", "bool"},
	}

	for _, gameVar := range gameVars {
		if err := fdm.AddVariable(gameVar.name, gameVar.simVar, gameVar.units); err != nil {
			log.Printf("⚠️  Warning: Failed to add game variable %s: %v", gameVar.name, err)
		}
	}

	// Start data collection
	if err := fdm.Start(); err != nil {
		log.Printf("❌ Failed to start data collection: %v", err)
		return
	}

	log.Println("✅ SimConnect connected successfully!")
	log.Println("🌤️  Weather data collection enabled")
	log.Println("🎮 Game information collection enabled")
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("📝 Debug: Received WebSocket connection request from %s", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("🔌 WebSocket client connected")

	ticker := time.NewTicker(500 * time.Millisecond) // 2Hz updates
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			data := collectFlightData()
			log.Printf("📤 Debug: Sending flight data via WebSocket to %s", conn.RemoteAddr())
			if err := conn.WriteJSON(data); err != nil {
				log.Printf("❌ WebSocket write error: %v", err)
				return
			}
		}
	}
}

func collectFlightData() FlightData {
	log.Printf("🔄 Debug: collectFlightData() called")
	data := FlightData{
		Connected: false,
		Error:     "",
	}

	// Check if SimConnect is available
	if simclient == nil || !simclient.IsOpen() {
		data.Error = "SimConnect not connected"
		log.Printf("❌ Debug: SimConnect check failed - simclient nil or not open")
		return data
	}

	if fdm == nil || !fdm.IsRunning() {
		data.Error = "Flight data manager not running"
		log.Printf("❌ Debug: FDM check failed - fdm nil or not running")
		return data
	}

	log.Printf("✅ Debug: SimConnect and FDM are both running")
	data.Connected = true
	// Collect all flight variables
	if alt, ok := fdm.GetVariable("Altitude"); ok {
		data.Altitude = alt.Value
		log.Printf("✈️ Debug: Altitude = %.0f ft", alt.Value)
	} else {
		log.Printf("✈️ Debug: Altitude variable not found")
	}
	if speed, ok := fdm.GetVariable("Indicated Airspeed"); ok {
		data.IndicatedSpeed = speed.Value
		log.Printf("🎯 Debug: Indicated Airspeed = %.1f kts", speed.Value)
	} else {
		log.Printf("🎯 Debug: Indicated Airspeed variable not found")
	}
	if ground, ok := fdm.GetVariable("Ground Speed"); ok {
		data.GroundSpeed = ground.Value
		log.Printf("🏃 Debug: Ground Speed = %.1f kts", ground.Value)
	} else {
		log.Printf("🏃 Debug: Ground Speed variable not found")
	}
	if vs, ok := fdm.GetVariable("Vertical Speed"); ok {
		data.VerticalSpeed = vs.Value
		log.Printf("📈 Debug: Vertical Speed = %.0f fpm", vs.Value)
	} else {
		log.Printf("📈 Debug: Vertical Speed variable not found")
	}
	if heading, ok := fdm.GetVariable("Heading Magnetic"); ok {
		data.HeadingMagnetic = heading.Value
		log.Printf("🧭 Debug: Heading Magnetic = %.1f°", heading.Value)
	} else {
		log.Printf("🧭 Debug: Heading Magnetic variable not found")
	}
	if bank, ok := fdm.GetVariable("Bank Angle"); ok {
		data.BankAngle = bank.Value
		log.Printf("🎢 Debug: Bank Angle = %.1f°", bank.Value)
	} else {
		log.Printf("🎢 Debug: Bank Angle variable not found")
	}
	if pitch, ok := fdm.GetVariable("Pitch Angle"); ok {
		data.PitchAngle = pitch.Value
		log.Printf("📐 Debug: Pitch Angle = %.1f°", pitch.Value)
	} else {
		log.Printf("📐 Debug: Pitch Angle variable not found")
	}
	if lat, ok := fdm.GetVariable("Latitude"); ok {
		data.Latitude = lat.Value
		log.Printf("🌍 Debug: Latitude = %.6f°", lat.Value)
	} else {
		log.Printf("🌍 Debug: Latitude variable not found")
	}
	if lon, ok := fdm.GetVariable("Longitude"); ok {
		data.Longitude = lon.Value
		log.Printf("🌎 Debug: Longitude = %.6f°", lon.Value)
	} else {
		log.Printf("🌎 Debug: Longitude variable not found")
	}

	// Engine and control variables
	if rpm, ok := fdm.GetVariable("Engine RPM"); ok {
		data.EngineRPM = rpm.Value
		log.Printf("🔥 Debug: Engine RPM = %.0f rpm", rpm.Value)
	} else {
		log.Printf("🔥 Debug: Engine RPM variable not found")
	}
	if throttle, ok := fdm.GetVariable("Throttle Position"); ok {
		data.ThrottlePos = throttle.Value
		log.Printf("🎛️ Debug: Throttle Position = %.1f%%", throttle.Value)
	} else {
		log.Printf("🎛️ Debug: Throttle Position variable not found")
	}
	if gear, ok := fdm.GetVariable("Gear Position"); ok {
		data.GearPosition = gear.Value
		log.Printf("⚙️ Debug: Gear Position = %.2f (interpreted as %s)", gear.Value, map[bool]string{true: "DOWN", false: "UP"}[gear.Value > 0.5])
	} else {
		log.Printf("⚙️ Debug: Gear Position variable not found")
	}
	if flaps, ok := fdm.GetVariable("Flaps Position"); ok {
		data.FlapsPosition = flaps.Value
		log.Printf("🛩️ Debug: Flaps Position = %.1f%%", flaps.Value)
	} else {
		log.Printf("🛩️ Debug: Flaps Position variable not found")
	}

	// Collect weather variables
	if temp, ok := fdm.GetVariable("Ambient Temperature"); ok {
		data.AmbientTemperature = temp.Value
		log.Printf("🌡️ Debug: Ambient Temperature = %.2f°C", temp.Value)
	} else {
		log.Printf("🌡️ Debug: Ambient Temperature variable not found")
	}
	if pressure, ok := fdm.GetVariable("Barometric Pressure"); ok {
		data.BarometricPressure = pressure.Value
		log.Printf("📊 Debug: Barometric Pressure = %.2f inHg", pressure.Value)
	} else {
		log.Printf("📊 Debug: Barometric Pressure variable not found")
	}
	if windSpeed, ok := fdm.GetVariable("Wind Speed"); ok {
		data.WindSpeed = windSpeed.Value
		log.Printf("💨 Debug: Wind Speed = %.2f kts", windSpeed.Value)
	} else {
		log.Printf("💨 Debug: Wind Speed variable not found")
	}
	if windDir, ok := fdm.GetVariable("Wind Direction"); ok {
		data.WindDirection = windDir.Value
		log.Printf("🧭 Debug: Wind Direction = %.2f°", windDir.Value)
	} else {
		log.Printf("🧭 Debug: Wind Direction variable not found")
	}
	if visibility, ok := fdm.GetVariable("Visibility"); ok {
		data.Visibility = visibility.Value
		log.Printf("👁️ Debug: Visibility = %.2f m", visibility.Value)
	} else {
		log.Printf("👁️ Debug: Visibility variable not found")
	}
	if clouds, ok := fdm.GetVariable("Cloud Coverage"); ok {
		data.CloudCoverage = clouds.Value
		log.Printf("☁️ Debug: Cloud Coverage = %.2f%%", clouds.Value)
	} else {
		log.Printf("☁️ Debug: Cloud Coverage variable not found")
	} // Collect game information variables
	if title, ok := fdm.GetVariable("Aircraft Title"); ok {
		// Convert float to string for aircraft title (may need special handling)
		data.AircraftTitle = fmt.Sprintf("Aircraft %.0f", title.Value)
		log.Printf("✈️ Debug: Aircraft Title = %.0f", title.Value)
	} else {
		log.Printf("✈️ Debug: Aircraft Title variable not found")
	}
	if simRate, ok := fdm.GetVariable("Simulation Rate"); ok {
		data.SimulationRate = simRate.Value
		log.Printf("⏱️ Debug: Simulation Rate = %.1fx", simRate.Value)
	} else {
		log.Printf("⏱️ Debug: Simulation Rate variable not found")
	}
	if localTime, ok := fdm.GetVariable("Local Time"); ok {
		// Convert seconds to time format
		hours := int(localTime.Value) / 3600
		minutes := (int(localTime.Value) % 3600) / 60
		data.LocalTime = fmt.Sprintf("%02d:%02d", hours, minutes)
		log.Printf("🕐 Debug: Local Time = %.0f seconds (%s)", localTime.Value, data.LocalTime)
	} else {
		log.Printf("🕐 Debug: Local Time variable not found")
	}
	if zuluTime, ok := fdm.GetVariable("Zulu Time"); ok {
		// Convert seconds to time format
		hours := int(zuluTime.Value) / 3600
		minutes := (int(zuluTime.Value) % 3600) / 60
		data.ZuluTime = fmt.Sprintf("%02d:%02dZ", hours, minutes)
		log.Printf("🌐 Debug: Zulu Time = %.0f seconds (%s)", zuluTime.Value, data.ZuluTime)
	} else {
		log.Printf("🌐 Debug: Zulu Time variable not found")
	}
	if onGround, ok := fdm.GetVariable("On Ground"); ok {
		data.OnGround = onGround.Value > 0.5
		log.Printf("🚁 Debug: On Ground = %.2f (interpreted as %t)", onGround.Value, data.OnGround)
	} else {
		log.Printf("🚁 Debug: On Ground variable not found")
	}
	if parkingBrake, ok := fdm.GetVariable("Parking Brake"); ok {
		data.ParkingBrake = parkingBrake.Value > 0.5
		log.Printf("🅿️ Debug: Parking Brake = %.2f (interpreted as %t)", parkingBrake.Value, data.ParkingBrake)
	} else {
		log.Printf("🅿️ Debug: Parking Brake variable not found")
	}
	if simPaused, ok := fdm.GetVariable("Sim Paused"); ok {
		data.SimPaused = simPaused.Value > 0.5
		log.Printf("⏸️ Debug: Sim Paused = %.2f (interpreted as %t)", simPaused.Value, data.SimPaused)
	} else {
		log.Printf("⏸️ Debug: Sim Paused variable not found")
	}

	// Get statistics
	dataCount, errorCount, lastUpdate := fdm.GetStats()
	data.DataCount = dataCount
	data.ErrorCount = errorCount
	data.LastUpdate = lastUpdate

	// Calculate update rate
	if elapsed := time.Since(startTime).Seconds(); elapsed > 0 {
		data.UpdateRate = float64(dataCount) / elapsed
	}

	return data
}
