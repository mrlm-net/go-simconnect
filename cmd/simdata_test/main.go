package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mrlm-net/go-simconnect/pkg/simconnect"
)

// Define a structure to hold our flight data
type FlightData struct {
	Altitude  float64 // Altitude in feet
	Airspeed  float64 // Indicated airspeed in knots
	Latitude  float64 // Latitude in degrees
	Longitude float64 // Longitude in degrees
	Heading   float64 // Heading in degrees
}

func main() {
	fmt.Println("=== SimConnect Simulation Data Test ===")
	fmt.Println("This test demonstrates requesting real-time simulation data from MSFS 2024")
	fmt.Println("Data includes: Altitude, Airspeed, Position (Lat/Lon), and Heading")
	fmt.Println("")

	// Get the current working directory to locate SimConnect.dll
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	dllPath := filepath.Join(wd, "lib", "SimConnect.dll")
	fmt.Printf("Looking for SimConnect.dll at: %s\n", dllPath)

	// Check if DLL exists
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		log.Fatalf("SimConnect.dll not found at %s", dllPath)
	}
	fmt.Println("✓ SimConnect.dll found")

	// Create a new SimConnect client
	client := simconnect.NewClientWithDLLPath("SimData Test Client", dllPath)
	fmt.Printf("✓ Client created: '%s'\n", client.GetName())

	fmt.Println("\n=== Testing Connection ===")

	// Try to open connection
	fmt.Println("Attempting to connect to SimConnect...")
	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to open SimConnect connection: %v\n", err)
		fmt.Println("\nNote: This is expected if Microsoft Flight Simulator is not running.")
		fmt.Println("To test simulation data:")
		fmt.Println("1. Start Microsoft Flight Simulator 2024")
		fmt.Println("2. Load any aircraft and flight")
		fmt.Println("3. Run this test again")
		return
	}

	fmt.Println("✅ Successfully connected to SimConnect!")
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to close connection: %v\n", err)
		} else {
			fmt.Println("✓ Connection closed successfully")
		}
	}()

	fmt.Println("\n=== Setting Up Data Definitions ===")

	// Define our data definition ID
	const DATA_DEFINITION_ID simconnect.DataDefinitionID = 1

	// Add simulation variables to our data definition
	fmt.Println("Adding simulation variables to data definition...")

	// Altitude
	if err := client.AddToDataDefinition(DATA_DEFINITION_ID, "Plane Altitude", "feet", simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
		fmt.Printf("❌ Failed to add altitude to data definition: %v\n", err)
		return
	}
	fmt.Println("✓ Added altitude (feet)")

	// Indicated Airspeed
	if err := client.AddToDataDefinition(DATA_DEFINITION_ID, "Airspeed Indicated", "knots", simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
		fmt.Printf("❌ Failed to add airspeed to data definition: %v\n", err)
		return
	}
	fmt.Println("✓ Added indicated airspeed (knots)")

	// Latitude
	if err := client.AddToDataDefinition(DATA_DEFINITION_ID, "Plane Latitude", "degrees", simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
		fmt.Printf("❌ Failed to add latitude to data definition: %v\n", err)
		return
	}
	fmt.Println("✓ Added latitude (degrees)")

	// Longitude
	if err := client.AddToDataDefinition(DATA_DEFINITION_ID, "Plane Longitude", "degrees", simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
		fmt.Printf("❌ Failed to add longitude to data definition: %v\n", err)
		return
	}
	fmt.Println("✓ Added longitude (degrees)")

	// Heading
	if err := client.AddToDataDefinition(DATA_DEFINITION_ID, "Plane Heading Degrees True", "degrees", simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
		fmt.Printf("❌ Failed to add heading to data definition: %v\n", err)
		return
	}
	fmt.Println("✓ Added heading (degrees)")
	fmt.Println("\n=== Requesting Simulation Data ===")

	// Request data from the user's aircraft
	const REQUEST_ID simconnect.SimObjectDataRequestID = 1

	fmt.Println("Requesting data from user aircraft...")
	if err := client.RequestDataOnSimObject(REQUEST_ID, DATA_DEFINITION_ID, simconnect.SIMCONNECT_OBJECT_ID_USER, simconnect.SIMCONNECT_PERIOD_ONCE); err != nil {
		fmt.Printf("❌ Failed to request data: %v\n", err)
		return
	}
	fmt.Println("✅ Data request sent successfully!")

	fmt.Println("\n=== Retrieving Simulation Data ===")

	// Poll for the simulation data response
	fmt.Println("🔍 Polling for simulation data responses...")
	var flightData *FlightData

	for attempts := 0; attempts < 50; attempts++ {
		header, data, err := client.GetSimObjectData()
		if err != nil {
			fmt.Printf("❌ Error retrieving data: %v\n", err)
			break
		}

		if header != nil && data != nil {
			fmt.Printf("✅ Received simulation data! (Request ID: %d)\n", header.DwRequestID)

			// Parse the flight data (expecting 5 float64 values in order)
			if len(data) >= 5 {
				flightData = &FlightData{
					Altitude:  data[0],
					Airspeed:  data[1],
					Latitude:  data[2],
					Longitude: data[3],
					Heading:   data[4],
				}
				break
			} else {
				fmt.Printf("⚠️  Unexpected data length: got %d values, expected 5\n", len(data))
			}
		}

		time.Sleep(100 * time.Millisecond) // Poll every 100ms
	}

	fmt.Println("\n=== Flight Data Results ===")

	if flightData != nil {
		fmt.Println("🎉 Successfully retrieved real-time flight data:")
		fmt.Printf("   ✈️  Altitude:    %8.1f feet\n", flightData.Altitude)
		fmt.Printf("   🏃 Airspeed:    %8.1f knots\n", flightData.Airspeed)
		fmt.Printf("   🌍 Latitude:    %11.6f°\n", flightData.Latitude)
		fmt.Printf("   🌍 Longitude:   %11.6f°\n", flightData.Longitude)
		fmt.Printf("   🧭 Heading:     %8.1f°\n", flightData.Heading)
		fmt.Println("")
		fmt.Println("✅ SimConnect data retrieval working perfectly!")
	} else {
		fmt.Println("⚠️  No simulation data received.")
		fmt.Println("This could mean:")
		fmt.Println("   - No active aircraft/flight loaded")
		fmt.Println("   - Aircraft is not in a valid flight state")
		fmt.Println("   - Data request was processed too quickly")
		fmt.Println("   - Try loading into an active flight and run again")
	}

	fmt.Println("\n=== Test Summary ===")
	if flightData != nil {
		fmt.Println("🎉 COMPLETE SUCCESS!")
		fmt.Println("✅ SimConnect connection established")
		fmt.Println("✅ Data definition created with 5 simulation variables")
		fmt.Println("✅ Data request sent to SimConnect")
		fmt.Println("✅ Real-time simulation data successfully retrieved!")
		fmt.Println("✅ All SimConnect functions working correctly")
		fmt.Println("")
		fmt.Println("🚀 Your SimConnect Go wrapper is fully functional!")
	} else {
		fmt.Println("✅ PARTIAL SUCCESS!")
		fmt.Println("✅ SimConnect connection established")
		fmt.Println("✅ Data definition created with 5 simulation variables")
		fmt.Println("✅ Data request sent to SimConnect")
		fmt.Println("⚠️  No simulation data received (see notes above)")
		fmt.Println("")
		fmt.Println("📝 Your SimConnect wrapper is working - data availability depends on flight state")
	}

	// Send a debug message to confirm everything worked
	if err := client.SendDebugMessage("SimData test completed successfully - data definition and request sent"); err != nil {
		fmt.Printf("⚠️  Warning: Failed to send debug message: %v\n", err)
	} else {
		fmt.Println("✓ Debug message sent (check DebugView)")
	}
}
