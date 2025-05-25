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
	fmt.Println("=== SimConnect Simulation Data DEBUG Test ===")
	fmt.Println("This debug version analyzes the received simulation data in detail")
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
	client := simconnect.NewClientWithDLLPath("SimData DEBUG Client", dllPath)
	fmt.Printf("✓ Client created: '%s'\n", client.GetName())

	fmt.Println("\n=== Testing Connection ===")

	// Try to open connection
	fmt.Println("Attempting to connect to SimConnect...")
	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to open SimConnect connection: %v\n", err)
		fmt.Println("\nNote: This is expected if Microsoft Flight Simulator is not running.")
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

	// Test individual simulation variables first
	testVariables := []struct {
		name     string
		simVar   string
		units    string
		dataType simconnect.SIMCONNECT_DATATYPE
	}{
		{"Altitude", "Plane Altitude", "feet", simconnect.SIMCONNECT_DATATYPE_FLOAT64},
		{"Airspeed", "Airspeed Indicated", "knots", simconnect.SIMCONNECT_DATATYPE_FLOAT64},
		{"Latitude", "Plane Latitude", "degrees", simconnect.SIMCONNECT_DATATYPE_FLOAT64},
		{"Longitude", "Plane Longitude", "degrees", simconnect.SIMCONNECT_DATATYPE_FLOAT64},
		{"Heading", "Plane Heading Degrees True", "degrees", simconnect.SIMCONNECT_DATATYPE_FLOAT64},
	}

	fmt.Println("Adding simulation variables to data definition...")

	for i, variable := range testVariables {
		fmt.Printf("  %d. Adding %s (%s in %s)...", i+1, variable.name, variable.simVar, variable.units)
		if err := client.AddToDataDefinition(DATA_DEFINITION_ID, variable.simVar, variable.units, variable.dataType); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			return
		}
		fmt.Println(" ✅ Success")
	}

	fmt.Println("\n=== Requesting Simulation Data ===")

	// Request data from the user's aircraft
	const REQUEST_ID simconnect.SimObjectDataRequestID = 1

	fmt.Println("Requesting data from user aircraft...")
	if err := client.RequestDataOnSimObject(REQUEST_ID, DATA_DEFINITION_ID, simconnect.SIMCONNECT_OBJECT_ID_USER, simconnect.SIMCONNECT_PERIOD_ONCE); err != nil {
		fmt.Printf("❌ Failed to request data: %v\n", err)
		return
	}
	fmt.Println("✅ Data request sent successfully!")

	fmt.Println("\n=== DEBUG: Raw Message Analysis ===")

	// Poll for ANY type of message to understand what we're getting
	fmt.Println("🔍 Polling for ANY SimConnect messages...")
	messageCount := 0

	for attempts := 0; attempts < 100; attempts++ {
		// Use the raw dispatch to see all message types
		data, err := client.GetRawDispatch()
		if err != nil {
			fmt.Printf("❌ Error retrieving message: %v\n", err)
			break
		}

		if data != nil {
			messageCount++

			// Parse message type
			msgType, err := simconnect.ParseMessageType(data)
			if err != nil {
				fmt.Printf("⚠️  Message %d: Failed to parse type: %v\n", messageCount, err)
				continue
			}

			fmt.Printf("📨 Message %d: Type=0x%08X, Size=%d bytes\n", messageCount, msgType, len(data))

			// Check specific message types
			switch msgType {
			case simconnect.SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
				fmt.Println("   🎯 This is SIMULATION OBJECT DATA!")

				// Parse simulation object data
				header, simData, err := simconnect.ParseSimObjectData(data)
				if err != nil {
					fmt.Printf("   ❌ Failed to parse simulation data: %v\n", err)
					continue
				}

				fmt.Printf("   📊 Header: RequestID=%d, ObjectID=%d, DefineID=%d\n",
					header.DwRequestID, header.DwObjectID, header.DwDefineID)
				fmt.Printf("   📊 Flags=%d, DefineCount=%d\n",
					header.DwFlags, header.DwDefineCount)

				if simData != nil {
					fmt.Printf("   📊 Data length: %d bytes\n", len(simData))

					// Show raw bytes
					fmt.Printf("   📊 Raw data (hex): ")
					for i, b := range simData {
						if i > 0 && i%8 == 0 {
							fmt.Printf(" ")
						}
						fmt.Printf("%02X", b)
						if i >= 63 { // Limit output
							fmt.Printf("...")
							break
						}
					}
					fmt.Println()

					// Try to parse as float64 values
					const float64Size = 8
					numFloats := len(simData) / float64Size
					fmt.Printf("   📊 Potential float64 values: %d\n", numFloats)

					if numFloats > 0 {
						fmt.Printf("   📊 Float64 values: ")
						for i := 0; i < numFloats && i < 10; i++ {
							offset := i * float64Size
							if offset+float64Size <= len(simData) {
								value := *(*float64)(unsafe.Pointer(&simData[offset]))
								fmt.Printf("%.3f ", value)
							}
						}
						fmt.Println()
					}
				} else {
					fmt.Println("   📊 No simulation data in message")
				}

			case simconnect.SIMCONNECT_RECV_ID_EXCEPTION:
				fmt.Println("   ⚠️  Exception message received")
			case simconnect.SIMCONNECT_RECV_ID_OPEN:
				fmt.Println("   🔗 Open confirmation message")
			case simconnect.SIMCONNECT_RECV_ID_QUIT:
				fmt.Println("   👋 Quit message")
			default:
				fmt.Printf("   📋 Other message type: 0x%08X\n", msgType)
			}

			fmt.Println()
		}

		time.Sleep(50 * time.Millisecond) // Poll every 50ms
	}

	fmt.Printf("\n=== Debug Summary ===\n")
	fmt.Printf("📊 Total messages received: %d\n", messageCount)

	if messageCount == 0 {
		fmt.Println("⚠️  No messages received - this could indicate:")
		fmt.Println("   - No active aircraft loaded")
		fmt.Println("   - Flight simulator not in flight mode")
		fmt.Println("   - Data requests processed too quickly")
	} else {
		fmt.Println("✅ Messages were received - check the analysis above!")
		fmt.Println("📝 Look for SIMULATION OBJECT DATA messages to see your flight data")
	}

	// Send a debug message
	if err := client.SendDebugMessage("SimData debug test completed"); err != nil {
		fmt.Printf("⚠️  Warning: Failed to send debug message: %v\n", err)
	} else {
		fmt.Println("✓ Debug message sent (check DebugView)")
	}
}
