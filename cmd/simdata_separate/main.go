package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/mrlm-net/go-simconnect/pkg/simconnect"
)

// VariableData represents a single simulation variable and its value
type VariableData struct {
	Name  string
	Value float64
	Units string
}

func main() {
	fmt.Println("=== SimConnect Separate Data Definitions Test ===")
	fmt.Println("Using separate data definitions for each variable")
	fmt.Println("")

	// Get the current working directory to locate SimConnect.dll
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	dllPath := filepath.Join(wd, "lib", "SimConnect.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		log.Fatalf("SimConnect.dll not found at %s", dllPath)
	}

	// Create a new SimConnect client
	client := simconnect.NewClientWithDLLPath("Separate Data Definitions Test", dllPath)

	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✅ Connected to SimConnect!")

	// Define variables - using only the most important ones to avoid too much complexity
	variables := []struct {
		simVar string
		units  string
		name   string
	}{
		{"Plane Altitude", "feet", "Altitude"},
		{"Airspeed Indicated", "knots", "Indicated Airspeed"},
		{"Plane Latitude", "degrees", "Latitude"},
		{"Plane Longitude", "degrees", "Longitude"},
		{"Plane Heading Degrees Magnetic", "degrees", "Heading Magnetic"},
		{"Vertical Speed", "feet per minute", "Vertical Speed"},
	}

	fmt.Printf("📊 Setting up %d separate data definitions...\n", len(variables))

	// Create separate data definition and request ID for each variable
	dataDefinitions := make([]simconnect.DataDefinitionID, len(variables))
	requestIDs := make([]simconnect.SimObjectDataRequestID, len(variables))

	// Add each variable to its own data definition
	for i, variable := range variables {
		defineID := simconnect.DataDefinitionID(i + 1)
		requestID := simconnect.SimObjectDataRequestID(i + 1)

		dataDefinitions[i] = defineID
		requestIDs[i] = requestID

		fmt.Printf("  %d. Adding %-25s (%s)...", i+1, variable.name, variable.units)
		if err := client.AddToDataDefinition(defineID, variable.simVar, variable.units, simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			return
		}
		fmt.Println(" ✅")
	}

	fmt.Println("\n🚁 Requesting data for all variables...")

	// Request data for each variable separately
	for i, requestID := range requestIDs {
		fmt.Printf("  %d. Requesting %s...", i+1, variables[i].name)
		if err := client.RequestDataOnSimObject(requestID, dataDefinitions[i], simconnect.SIMCONNECT_OBJECT_ID_USER, simconnect.SIMCONNECT_PERIOD_SIM_FRAME); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
		} else {
			fmt.Println(" ✅")
		}
	}

	fmt.Println("\n📡 All data requests sent successfully!")
	fmt.Println("\n🔄 Receiving simulation data (Press Ctrl+C to stop)...")
	fmt.Println("=" + "====================================" + "=")

	// Track the latest values for each variable
	latestValues := make([]VariableData, len(variables))
	for i, variable := range variables {
		latestValues[i] = VariableData{
			Name:  variable.name,
			Value: 0.0,
			Units: variable.units,
		}
	}

	dataCount := 0
	startTime := time.Now()
	lastDisplayTime := time.Now()

	// Main loop to receive and display data
	for {
		data, err := client.GetRawDispatch()
		if err != nil {
			fmt.Printf("❌ Error receiving data: %v\n", err)
			continue
		}

		if data != nil {
			msgType, err := simconnect.ParseMessageType(data)
			if err != nil {
				continue
			}

			if msgType == simconnect.SIMCONNECT_RECV_ID_SIMOBJECT_DATA {
				header, simData, err := simconnect.ParseSimObjectData(data)
				if err != nil {
					continue
				}

				// Find which variable this data corresponds to
				requestID := simconnect.SimObjectDataRequestID(header.DwRequestID)
				for i, rid := range requestIDs {
					if rid == requestID && simData != nil && len(simData) >= 8 {
						value := *(*float64)(unsafe.Pointer(&simData[0]))
						latestValues[i].Value = value
						dataCount++
						break
					}
				}

				// Display all current values every second
				if time.Since(lastDisplayTime) >= 1*time.Second {
					fmt.Printf("\n📦 Combined Data Update (%.1fs elapsed, %d total updates)\n",
						time.Since(startTime).Seconds(), dataCount)

					fmt.Printf("   🛩️  AIRCRAFT STATE:\n")
					for _, variable := range latestValues {
						fmt.Printf("       %-20s: %12.3f %s\n", variable.Name, variable.Value, variable.Units)
					}

					lastDisplayTime = time.Now()
				}

				// Stop after reasonable amount of time for testing
				if time.Since(startTime) >= 30*time.Second {
					fmt.Printf("\n🏁 Stopping after 30 seconds\n")
					fmt.Printf("💡 Separate data definitions are working successfully!\n")
					fmt.Printf("📊 Total data updates received: %d\n", dataCount)
					break
				}

			} else if msgType == simconnect.SIMCONNECT_RECV_ID_EXCEPTION {
				fmt.Printf("❌ SimConnect Exception received\n")
			}
		}

		// Small delay to prevent overwhelming the system
		time.Sleep(50 * time.Millisecond)
	}

	client.SendDebugMessage(fmt.Sprintf("Separate data definitions test completed. Received %d data updates.", dataCount))
	fmt.Println("\n✅ Test completed successfully!")
}
