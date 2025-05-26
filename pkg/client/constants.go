package client

// System state constants for RequestSystemState function
// These match the values documented in the SimConnect API reference
const (
	// SystemStateAircraftLoaded requests the full path name of the last loaded aircraft flight dynamics file (.AIR extension)
	SystemStateAircraftLoaded = "AircraftLoaded"

	// SystemStateDialogMode requests whether the simulation is in Dialog mode or not
	SystemStateDialogMode = "DialogMode"

	// SystemStateFlightLoaded requests the full path name of the last loaded flight (.FLT extension)
	SystemStateFlightLoaded = "FlightLoaded"

	// SystemStateFlightPlan requests the full path name of the active flight plan (empty string if none active)
	SystemStateFlightPlan = "FlightPlan"

	// SystemStateSim requests the state of the simulation (1 = user in control, 0 = navigating UI)
	SystemStateSim = "Sim"
)

// SimConnect data types for variable definitions
type SIMCONNECT_DATATYPE uint32

const (
	SIMCONNECT_DATATYPE_INVALID   SIMCONNECT_DATATYPE = 0
	SIMCONNECT_DATATYPE_INT32     SIMCONNECT_DATATYPE = 1
	SIMCONNECT_DATATYPE_INT64     SIMCONNECT_DATATYPE = 2
	SIMCONNECT_DATATYPE_FLOAT32   SIMCONNECT_DATATYPE = 3
	SIMCONNECT_DATATYPE_FLOAT64   SIMCONNECT_DATATYPE = 4
	SIMCONNECT_DATATYPE_STRING8   SIMCONNECT_DATATYPE = 5
	SIMCONNECT_DATATYPE_STRING32  SIMCONNECT_DATATYPE = 6
	SIMCONNECT_DATATYPE_STRING64  SIMCONNECT_DATATYPE = 7
	SIMCONNECT_DATATYPE_STRING128 SIMCONNECT_DATATYPE = 8
	SIMCONNECT_DATATYPE_STRING256 SIMCONNECT_DATATYPE = 9
	SIMCONNECT_DATATYPE_STRING260 SIMCONNECT_DATATYPE = 10
	SIMCONNECT_DATATYPE_STRINGV   SIMCONNECT_DATATYPE = 11
)

// SimConnect data request periods
type SIMCONNECT_PERIOD uint32

const (
	SIMCONNECT_PERIOD_NEVER        SIMCONNECT_PERIOD = 0
	SIMCONNECT_PERIOD_ONCE         SIMCONNECT_PERIOD = 1
	SIMCONNECT_PERIOD_VISUAL_FRAME SIMCONNECT_PERIOD = 2
	SIMCONNECT_PERIOD_SIM_FRAME    SIMCONNECT_PERIOD = 3
	SIMCONNECT_PERIOD_SECOND       SIMCONNECT_PERIOD = 4
)

// SimConnect data request flags
type SIMCONNECT_DATA_REQUEST_FLAG uint32

const (
	SIMCONNECT_DATA_REQUEST_FLAG_DEFAULT SIMCONNECT_DATA_REQUEST_FLAG = 0
	SIMCONNECT_DATA_REQUEST_FLAG_CHANGED SIMCONNECT_DATA_REQUEST_FLAG = 1
	SIMCONNECT_DATA_REQUEST_FLAG_TAGGED  SIMCONNECT_DATA_REQUEST_FLAG = 2
)

// SimConnect object IDs
type SIMCONNECT_OBJECT_ID uint32

const (
	SIMCONNECT_OBJECT_ID_USER SIMCONNECT_OBJECT_ID = 0
)

// Data definition and request ID types
type DataDefinitionID uint32
type SimObjectDataRequestID uint32

// SimConnect data set flags for SetDataOnSimObject
type SIMCONNECT_DATA_SET_FLAG uint32

const (
	SIMCONNECT_DATA_SET_FLAG_DEFAULT SIMCONNECT_DATA_SET_FLAG = 0
	SIMCONNECT_DATA_SET_FLAG_TAGGED  SIMCONNECT_DATA_SET_FLAG = 1
)

// SimConnect client event ID type for system events
type SIMCONNECT_CLIENT_EVENT_ID uint32

// SimConnect system event state
type SIMCONNECT_STATE uint32

const (
	SIMCONNECT_STATE_OFF SIMCONNECT_STATE = 0
	SIMCONNECT_STATE_ON  SIMCONNECT_STATE = 1
)

// System Event Names - exact strings from Microsoft Flight Simulator 2024 documentation
// These are case-insensitive according to the docs but we use exact casing for consistency
const (
	// Timer Events
	SystemEvent1Sec  = "1sec"  // Request a notification every second
	SystemEvent4Sec  = "4sec"  // Request a notification every four seconds
	SystemEvent6Hz   = "6Hz"   // Request notifications six times per second (joystick rate)
	SystemEventFrame = "Frame" // Request notifications every visual frame

	// Flight and Aircraft Events
	SystemEventAircraftLoaded = "AircraftLoaded" // When aircraft flight dynamics file is changed (.AIR extension)
	SystemEventFlightLoaded   = "FlightLoaded"   // When a flight is loaded (includes filename)
	SystemEventFlightSaved    = "FlightSaved"    // When a flight is saved correctly (includes filename)

	// Flight Plan Events
	SystemEventFlightPlanActivated   = "FlightPlanActivated"   // When a new flight plan is activated (includes filename)
	SystemEventFlightPlanDeactivated = "FlightPlanDeactivated" // When the active flight plan is deactivated

	// Simulation State Events
	SystemEventSim      = "Sim"      // Simulation running state (1 = running, 0 = not running)
	SystemEventSimStart = "SimStart" // The simulator is running (user actively controlling)
	SystemEventSimStop  = "SimStop"  // The simulator is not running (loading, UI navigation)

	// Pause Events
	SystemEventPause      = "Pause"      // Flight paused/unpaused (1 = paused, 0 = unpaused)
	SystemEventPauseEx    = "Pause_EX1"  // Extended pause state with detailed flags
	SystemEventPaused     = "Paused"     // Notification when flight is paused
	SystemEventUnpaused   = "Unpaused"   // Notification when flight is unpaused
	SystemEventPauseFrame = "PauseFrame" // Every visual frame while simulation is paused

	// Crash Events
	SystemEventCrashed    = "Crashed"    // User aircraft crashes
	SystemEventCrashReset = "CrashReset" // Crash cut-scene has completed

	// Object Events
	SystemEventObjectAdded   = "ObjectAdded"   // AI object added to simulation
	SystemEventObjectRemoved = "ObjectRemoved" // AI object removed from simulation

	// Position and View Events
	SystemEventPositionChanged = "PositionChanged" // User changes aircraft position through dialog
	SystemEventView            = "View"            // User aircraft view is changed

	// System Events
	SystemEventSound = "Sound" // Master sound switch changed (0 = off, 1 = on)

	// Legacy Events (kept for compatibility)
	SystemEventCustomMissionActionExecuted = "CustomMissionActionExecuted" // Legacy mission action executed
	SystemEventWeatherModeChanged          = "WeatherModeChanged"          // Legacy weather mode changed
)

// Pause state flags for Pause_EX1 event (from Microsoft documentation)
const (
	PAUSE_STATE_FLAG_OFF              = 0 // No Pause
	PAUSE_STATE_FLAG_PAUSE            = 1 // "full" Pause (sim + traffic + etc...)
	PAUSE_STATE_FLAG_PAUSE_WITH_SOUND = 2 // FSX Legacy Pause (not used anymore)
	PAUSE_STATE_FLAG_ACTIVE_PAUSE     = 4 // Pause was activated using the "Active Pause" Button
	PAUSE_STATE_FLAG_SIM_PAUSE        = 8 // Pause the player sim but traffic, multi, etc... will still run
)

// View system event data flags (from Microsoft documentation)
const (
	SIMCONNECT_VIEW_SYSTEM_EVENT_DATA_COCKPIT_2D      = 0x00000001 // 2D Cockpit view
	SIMCONNECT_VIEW_SYSTEM_EVENT_DATA_COCKPIT_VIRTUAL = 0x00000002 // Virtual Cockpit view
	SIMCONNECT_VIEW_SYSTEM_EVENT_DATA_ORTHOGONAL      = 0x00000003 // Map view
)

// Sound system event data flags (from Microsoft documentation)
const (
	SIMCONNECT_SOUND_SYSTEM_EVENT_DATA_MASTER = 0x00000001 // Master sound is on
)
