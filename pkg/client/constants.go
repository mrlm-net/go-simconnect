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
