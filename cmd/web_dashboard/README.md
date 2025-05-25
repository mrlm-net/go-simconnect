# MSFS 2024 Web Dashboard

Beautiful real-time flight dashboard for Microsoft Flight Simulator 2024 with modern web UI.

## Features

🎨 **Modern UI Design**
- Beautiful Tailwind CSS styling with gradient cards
- Dark theme optimized for flight simulation
- Responsive design for desktop and mobile
- Real-time status indicators and animations

📊 **Real-time Flight Data**
- Primary flight instruments (altitude, airspeed, heading)
- Aircraft attitude and engine parameters
- Landing gear and flap positions
- Geographic coordinates and navigation data
- Performance statistics and error monitoring

🌤️ **Weather Information**
- Ambient temperature and barometric pressure
- Wind speed and direction
- Visibility distance
- Cloud coverage percentage

🎮 **Game Information**
- Aircraft type and identification
- Simulation rate and pause status monitoring
- Local and UTC (Zulu) time
- Ground status and parking brake

🔌 **WebSocket Integration**
- Real-time data streaming at 2Hz
- Automatic reconnection on connection loss
- Connection status monitoring
- Error handling and recovery

## Quick Start

### 1. Prerequisites
- MSFS 2024 running with SimConnect enabled
- Go 1.19+ installed
- Modern web browser (Chrome, Firefox, Edge, Safari)

### 2. Build and Run
```bash
# Build the web dashboard
go build -o cmd/web_dashboard/dashboard.exe cmd/web_dashboard/main.go

# Run the dashboard server
./cmd/web_dashboard/dashboard.exe
```

### 3. Access Dashboard
Open your web browser and navigate to:
```
http://localhost:8080
```

## Dashboard Layout

### Primary Cards
- **🛫 ALTITUDE** - Current altitude and vertical speed
- **⚡ AIRSPEED** - Indicated and ground speed
- **🧭 HEADING** - Magnetic heading

### Secondary Panels
- **📐 Attitude** - Bank and pitch angles
- **🔥 Engine** - RPM and throttle position  
- **🎛️ Controls** - Gear and flap positions
- **📍 Position** - Latitude and longitude coordinates

### Weather Panel
- **🌡️ Temperature** - Ambient temperature in Celsius
- **📊 Pressure** - Barometric pressure in millibars
- **💨 Wind** - Wind speed (knots) and direction (degrees)
- **👁️ Visibility** - Visibility distance in meters
- **☁️ Clouds** - Total cloud coverage percentage

### Aircraft Panel
- **✈️ Aircraft** - Current aircraft type
- **🚁 Ground Status** - On ground indicator
- **🅿️ Parking Brake** - Parking brake status

### Simulation Panel
- **⏱️ Sim Rate** - Current simulation rate multiplier
- **⏸️ Status** - Simulation pause state (RUNNING/PAUSED)
- **🕐 Local Time** - Local simulation time
- **🌐 Zulu Time** - UTC time in simulation

### Statistics Panel
- Data collection rate and total data points
- Error count and monitoring
- Last update timestamp

## Configuration

The dashboard uses the same SimConnect configuration as other demos:
- **Default DLL Path**: `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
- **Server Port**: `8080` (configurable in source)
- **Update Rate**: `2Hz` WebSocket updates (configurable in source)

## Customization

### Changing Update Rate
Edit `main.go` line ~90:
```go
ticker := time.NewTicker(500 * time.Millisecond) // 2Hz updates
```

### Changing Port
Edit `main.go` line ~54:
```go
log.Fatal(http.ListenAndServe(":8080", nil))
```

### Adding Custom Variables
The dashboard uses the standard 15 flight variables. To add custom variables, modify the `initSimConnect()` function in `main.go`:

```go
// Add custom variables after AddStandardVariables()
fdm.AddVariable("Fuel Quantity", "Fuel Total Quantity", "gallons")
```

Then update the `collectFlightData()` function to include the new variable:
```go
if fuel, ok := fdm.GetVariable("Fuel Quantity"); ok {
    data.FuelQuantity = fuel.Value
}
```

## Troubleshooting

### Dashboard Not Loading
- Check that MSFS 2024 is running and fully loaded
- Verify SimConnect is enabled in MSFS General Options > Developers
- Ensure port 8080 is not blocked by firewall
- Check console output for connection errors

### No Flight Data
- Ensure aircraft is loaded (not in main menu)
- Check WebSocket connection in browser developer tools
- Verify SimConnect.dll path is correct
- Check error messages in dashboard statistics panel

### Performance Issues
- Monitor update rate in statistics panel
- Check MSFS frame rate and system performance
- Reduce update frequency if needed
- Close other SimConnect applications

## Architecture

```
┌─────────────────┐    WebSocket    ┌─────────────────┐
│   Web Browser   │ ◄──────────────► │   Go Server     │
│                 │     (2Hz)       │                 │
│ • Tailwind CSS  │                 │ • HTTP Server   │
│ • JavaScript    │                 │ • WebSocket     │
│ • Auto-reconnect│                 │ • JSON API      │
└─────────────────┘                 └─────────────────┘
                                             │
                                             │ SimConnect API
                                             ▼
                                    ┌─────────────────┐
                                    │ MSFS 2024       │
                                    │ SimConnect.dll  │
                                    └─────────────────┘
```

## Development

The web dashboard is built with:
- **Backend**: Go standard library + gorilla/websocket
- **Frontend**: HTML5 + Tailwind CSS + Vanilla JavaScript
- **Real-time**: WebSocket for live data streaming
- **Styling**: Tailwind CSS via CDN (no build step required)

Files:
- `main.go` - Go HTTP server and WebSocket handler
- `static/index.html` - Dashboard HTML and UI
- `static/app.js` - WebSocket client and data visualization

## Browser Compatibility

✅ **Supported Browsers**
- Chrome 60+
- Firefox 55+
- Safari 11+
- Edge 79+

⚠️ **WebSocket Support Required**
- All modern browsers support WebSockets
- Internet Explorer not supported
