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

func main() {
	fmt.Println("=== SimConnect Individual Variable Test ===")
	fmt.Println("Testing each simulation variable individually to identify working ones")
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
	client := simconnect.NewClientWithDLLPath("Individual Variable Test", dllPath)

	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✅ Connected to SimConnect!")

	// Test variables individually
	testVariables := []struct {
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

	successfulVars := []struct {
		name  string
		value float64
	}{}

	for i, variable := range testVariables {
		fmt.Printf("\n=== Test %d: %s ===\n", i+1, variable.name)
		fmt.Printf("Variable: %s (%s)\n", variable.simVar, variable.units)

		// Create unique IDs for each test
		defineID := simconnect.DataDefinitionID(i + 1)
		requestID := simconnect.SimObjectDataRequestID(i + 1)

		// Add this variable to its own data definition
		fmt.Print("Adding to data definition...")
		if err := client.AddToDataDefinition(defineID, variable.simVar, variable.units, simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			continue
		}
		fmt.Println(" ✅ Success")

		// Request data for this variable
		fmt.Print("Requesting data...")
		if err := client.RequestDataOnSimObject(requestID, defineID, simconnect.SIMCONNECT_OBJECT_ID_USER, simconnect.SIMCONNECT_PERIOD_ONCE); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			continue
		}
		fmt.Println(" ✅ Success")

		// Poll for response
		fmt.Print("Polling for response...")
		dataReceived := false

		for attempts := 0; attempts < 20; attempts++ {
			data, err := client.GetRawDispatch()
			if err != nil {
				fmt.Printf(" ❌ Error: %v\n", err)
				break
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

					if header.DwRequestID == uint32(requestID) && simData != nil && len(simData) >= 8 {
						value := *(*float64)(unsafe.Pointer(&simData[0]))
						fmt.Printf(" ✅ SUCCESS: %.3f %s\n", value, variable.units)
						successfulVars = append(successfulVars, struct {
							name  string
							value float64
						}{variable.name, value})
						dataReceived = true
						break
					}
				} else if msgType == simconnect.SIMCONNECT_RECV_ID_EXCEPTION {
					fmt.Printf(" ❌ Exception (variable not available)\n")
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
		}

		if !dataReceived {
			fmt.Printf(" ⚠️  No data received\n")
		}

		// Small delay between tests
		time.Sleep(200 * time.Millisecond)
	}
	// Summary
	separator := "============================================================"
	fmt.Println("\n" + separator)
	fmt.Println("=== SUMMARY OF WORKING VARIABLES ===")
	fmt.Println(separator)

	if len(successfulVars) == 0 {
		fmt.Println("❌ No variables returned data successfully")
		fmt.Println("This could mean:")
		fmt.Println("   - No aircraft is loaded")
		fmt.Println("   - Not in an active flight")
		fmt.Println("   - Aircraft doesn't support these variables")
	} else {
		fmt.Printf("✅ Successfully retrieved %d variables:\n\n", len(successfulVars))

		for i, variable := range successfulVars {
			fmt.Printf("%2d. %-20s: %12.3f\n", i+1, variable.name, variable.value)
		}

		fmt.Println("\n🎉 These variables can be used together in a combined data definition!")
		fmt.Println("💡 Use only the working variables for reliable data retrieval.")
	}

	client.SendDebugMessage(fmt.Sprintf("Individual variable test completed. %d/%d variables successful.", len(successfulVars), len(testVariables)))
}
