# SimConnect Variables Reference

This document provides a reference for commonly used Microsoft Flight Simulator variables that can be accessed through the go-simconnect library.

## Variable Categories

### Aircraft Position and Attitude

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `PLANE LATITUDE` | degrees | Aircraft latitude position | ❌ |
| `PLANE LONGITUDE` | degrees | Aircraft longitude position | ❌ |
| `PLANE ALTITUDE` | feet | Aircraft altitude above sea level | ❌ |
| `INDICATED ALTITUDE` | feet | Altimeter reading | ❌ |
| `PLANE HEADING DEGREES TRUE` | degrees | True heading | ❌ |
| `PLANE HEADING DEGREES MAGNETIC` | degrees | Magnetic heading | ❌ |
| `PLANE PITCH DEGREES` | degrees | Aircraft pitch attitude | ❌ |
| `PLANE BANK DEGREES` | degrees | Aircraft bank angle | ❌ |

### Flight Dynamics

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `AIRSPEED INDICATED` | knots | Indicated airspeed | ❌ |
| `AIRSPEED TRUE` | knots | True airspeed | ❌ |
| `GROUND VELOCITY` | knots | Ground speed | ❌ |
| `VERTICAL SPEED` | feet per minute | Rate of climb/descent | ❌ |
| `ACCELERATION BODY X` | feet per second squared | Longitudinal acceleration | ❌ |
| `ACCELERATION BODY Y` | feet per second squared | Lateral acceleration | ❌ |
| `ACCELERATION BODY Z` | feet per second squared | Vertical acceleration | ❌ |

### Engine and Systems

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `GENERAL ENG RPM:1` | rpm | Engine 1 RPM | ❌ |
| `ENG THROTTLE LEVER POSITION:1` | percent | Engine 1 throttle position | ✅ |
| `PROP RPM:1` | rpm | Propeller 1 RPM | ❌ |
| `ENG MANIFOLD PRESSURE:1` | inHg | Engine 1 manifold pressure | ❌ |
| `ENG FUEL FLOW GPH:1` | gallons per hour | Engine 1 fuel flow | ❌ |
| `FUEL TANK LEFT MAIN QUANTITY` | gallons | Left main tank fuel quantity | ❌ |
| `FUEL TANK RIGHT MAIN QUANTITY` | gallons | Right main tank fuel quantity | ❌ |

### Flight Controls

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `ELEVATOR POSITION` | percent | Elevator deflection | ❌ |
| `AILERON POSITION` | percent | Aileron deflection | ❌ |
| `RUDDER POSITION` | percent | Rudder deflection | ❌ |
| `FLAPS HANDLE PERCENT` | percent | Flaps handle position | ✅ |
| `SPOILERS HANDLE POSITION` | percent | Spoilers handle position | ✅ |
| `GEAR HANDLE POSITION` | bool | Landing gear handle position | ✅ |

### Autopilot

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `AUTOPILOT MASTER` | bool | Autopilot master switch | ✅ |
| `AUTOPILOT HEADING LOCK` | bool | Heading hold mode | ✅ |
| `AUTOPILOT ALTITUDE LOCK` | bool | Altitude hold mode | ✅ |
| `AUTOPILOT AIRSPEED HOLD` | bool | Airspeed hold mode | ✅ |
| `AUTOPILOT HEADING LOCK DIR` | degrees | Target heading | ✅ |
| `AUTOPILOT ALTITUDE LOCK VAR` | feet | Target altitude | ✅ |
| `AUTOPILOT AIRSPEED HOLD VAR` | knots | Target airspeed | ✅ |

### Weather and Environment

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `AMBIENT WIND VELOCITY` | knots | Wind speed | ❌ |
| `AMBIENT WIND DIRECTION` | degrees | Wind direction | ❌ |
| `AMBIENT TEMPERATURE` | celsius | Outside air temperature | ❌ |
| `BAROMETER PRESSURE` | millibars | Barometric pressure | ❌ |
| `SEA LEVEL PRESSURE` | millibars | Sea level pressure | ❌ |

### Camera and View

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `CAMERA STATE` | number | Current camera/view state | ✅ |
| `CAMERA SUBSTATE` | number | Camera sub-state | ✅ |

#### Camera State Values

| Value | Description |
|-------|-------------|
| 2 | Cockpit View |
| 3 | External View |
| 4 | Wing View |
| 5 | Tail View |
| 6 | Tower View |
| 10 | Instrument View |

### Navigation

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `NAV OBS:1` | degrees | VOR 1 OBS setting | ✅ |
| `NAV OBS:2` | degrees | VOR 2 OBS setting | ✅ |
| `NAV RADIAL:1` | degrees | VOR 1 radial | ❌ |
| `NAV RADIAL:2` | degrees | VOR 2 radial | ❌ |
| `GPS GROUND SPEED` | knots | GPS ground speed | ❌ |
| `GPS GROUND TRUE TRACK` | degrees | GPS ground track | ❌ |

### Communication

| Variable Name | Units | Description | Writable |
|---------------|-------|-------------|----------|
| `COM ACTIVE FREQUENCY:1` | MHz | COM 1 active frequency | ✅ |
| `COM STANDBY FREQUENCY:1` | MHz | COM 1 standby frequency | ✅ |
| `COM ACTIVE FREQUENCY:2` | MHz | COM 2 active frequency | ✅ |
| `COM STANDBY FREQUENCY:2` | MHz | COM 2 standby frequency | ✅ |

## Usage Examples

### Basic Flight Data

```go
// Add essential flight variables
fdm.AddVariable("Airspeed", "AIRSPEED INDICATED", "knots")
fdm.AddVariable("Altitude", "INDICATED ALTITUDE", "feet")
fdm.AddVariable("Heading", "PLANE HEADING DEGREES MAGNETIC", "degrees")
fdm.AddVariable("Vertical Speed", "VERTICAL SPEED", "feet per minute")
```

### Engine Monitoring

```go
// Add engine parameters
fdm.AddVariable("Engine RPM", "GENERAL ENG RPM:1", "rpm")
fdm.AddVariable("Manifold Pressure", "ENG MANIFOLD PRESSURE:1", "inHg")
fdm.AddVariable("Fuel Flow", "ENG FUEL FLOW GPH:1", "gallons per hour")
fdm.AddVariable("Left Tank", "FUEL TANK LEFT MAIN QUANTITY", "gallons")
```

### Autopilot Control

```go
// Add autopilot variables with write capability
fdm.AddVariableWithWritable("AP Master", "AUTOPILOT MASTER", "bool", true)
fdm.AddVariableWithWritable("AP Heading", "AUTOPILOT HEADING LOCK", "bool", true)
fdm.AddVariableWithWritable("AP Target Heading", "AUTOPILOT HEADING LOCK DIR", "degrees", true)

// Set autopilot values
fdm.SetVariable("AP Master", 1)      // Enable autopilot
fdm.SetVariable("AP Heading", 1)     // Enable heading hold
fdm.SetVariable("AP Target Heading", 90) // Set heading to 090°
```

### Camera Control

```go
// Add camera control
fdm.AddVariableWithWritable("Camera", "CAMERA STATE", "number", true)

// Cycle through views
fdm.SetVariable("Camera", 2)  // Cockpit
fdm.SetVariable("Camera", 3)  // External
fdm.SetVariable("Camera", 4)  // Wing
fdm.SetVariable("Camera", 6)  // Tower
```

## Important Notes

### Units

- Always use the exact unit strings as specified in the table
- Case matters for both variable names and units
- Use "bool" for boolean values (0 = false, 1 = true)

### Indexed Variables

Some variables support multiple instances using the `:index` syntax:
- `:1` refers to the first instance (engines, radios, etc.)
- `:2` refers to the second instance
- Not all aircraft support multiple instances

### Writable Variables

Variables marked as ✅ Writable can be modified using `SetVariable()`. Always check that:
1. The variable is actually writable in the simulator
2. The aircraft supports the specific system/control
3. The value is within valid ranges for that variable

### Error Handling

Some variables may not be available depending on:
- Aircraft type (jets vs. props vs. helicopters)
- Aircraft complexity (study-level vs. basic)
- Simulator state (on ground vs. in flight)
- System failures or damage

Always handle errors when adding variables or setting values.

## Finding More Variables

This list covers common variables. For a complete reference:
1. Consult the Microsoft Flight Simulator SDK documentation
2. Use tools like FSUIPC or other SimConnect utilities to explore available variables
3. Check aircraft-specific documentation for custom variables
4. Experiment with similar variable names following SimConnect naming conventions
