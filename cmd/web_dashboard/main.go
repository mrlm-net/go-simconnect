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
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/web_dashboard/static/"))))

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

	// Add standard variables
	if err := fdm.AddStandardVariables(); err != nil {
		log.Printf("❌ Failed to add variables: %v", err)
		return
	}

	// Start data collection
	if err := fdm.Start(); err != nil {
		log.Printf("❌ Failed to start data collection: %v", err)
		return
	}

	log.Println("✅ SimConnect connected successfully!")
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "cmd/web_dashboard/static/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
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
			if err := conn.WriteJSON(data); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func collectFlightData() FlightData {
	data := FlightData{
		Connected: false,
		Error:     "",
	}

	// Check if SimConnect is available
	if simclient == nil || !simclient.IsOpen() {
		data.Error = "SimConnect not connected"
		return data
	}

	if fdm == nil || !fdm.IsRunning() {
		data.Error = "Flight data manager not running"
		return data
	}

	data.Connected = true

	// Collect all flight variables
	if alt, ok := fdm.GetVariable("Altitude"); ok {
		data.Altitude = alt.Value
	}
	if speed, ok := fdm.GetVariable("Indicated Airspeed"); ok {
		data.IndicatedSpeed = speed.Value
	}
	if ground, ok := fdm.GetVariable("Ground Speed"); ok {
		data.GroundSpeed = ground.Value
	}
	if vs, ok := fdm.GetVariable("Vertical Speed"); ok {
		data.VerticalSpeed = vs.Value
	}
	if heading, ok := fdm.GetVariable("Heading Magnetic"); ok {
		data.HeadingMagnetic = heading.Value
	}
	if bank, ok := fdm.GetVariable("Bank Angle"); ok {
		data.BankAngle = bank.Value
	}
	if pitch, ok := fdm.GetVariable("Pitch Angle"); ok {
		data.PitchAngle = pitch.Value
	}
	if rpm, ok := fdm.GetVariable("Engine RPM"); ok {
		data.EngineRPM = rpm.Value
	}
	if throttle, ok := fdm.GetVariable("Throttle Position"); ok {
		data.ThrottlePos = throttle.Value
	}
	if gear, ok := fdm.GetVariable("Gear Position"); ok {
		data.GearPosition = gear.Value
	}
	if flaps, ok := fdm.GetVariable("Flaps Position"); ok {
		data.FlapsPosition = flaps.Value
	}
	if lat, ok := fdm.GetVariable("Latitude"); ok {
		data.Latitude = lat.Value
	}
	if lon, ok := fdm.GetVariable("Longitude"); ok {
		data.Longitude = lon.Value
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
