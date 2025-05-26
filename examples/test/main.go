package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() {
	fmt.Println("=== SimConnect Go Client Test ===")

	// Use the MSFS 2024 SDK SimConnect.dll
	dllPath := "C:\\MSFS 2024 SDK\\SimConnect SDK\\lib\\SimConnect.dll"
	fmt.Printf("Looking for SimConnect.dll at: %s\n", dllPath)

	// Check if DLL exists
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		log.Fatalf("SimConnect.dll not found at %s", dllPath)
	}
	fmt.Println("✓ SimConnect.dll found")
	// Create a new SimConnect client with the local DLL
	simclient := client.NewClientWithDLLPath("Go Test Client", dllPath)
	fmt.Printf("✓ Client created: '%s'\n", simclient.GetName())
	fmt.Printf("✓ Initial connection state: %v\n", simclient.IsOpen())

	fmt.Println("\n=== Testing Connection ===")
	// Try to open connection
	fmt.Println("Attempting to connect to SimConnect...")
	if err := simclient.Open(); err != nil {
		// This might fail if MSFS is not running, which is expected
		fmt.Printf("❌ Failed to open SimConnect connection: %v\n", err)
		fmt.Println("\nNote: This is expected if Microsoft Flight Simulator is not running.")
		fmt.Println("To test successfully:")
		fmt.Println("1. Start Microsoft Flight Simulator 2024")
		fmt.Println("2. Load into a flight or main menu")
		fmt.Println("3. Run this test again")
		return
	}

	fmt.Printf("✅ Successfully connected to SimConnect!\n")
	fmt.Printf("✓ Connection handle: 0x%X\n", simclient.GetHandle())
	fmt.Printf("✓ Connection state: %v\n", simclient.IsOpen())
	fmt.Println("\n=== Testing System State Requests ===")

	// Test requesting system states
	testRequests := []struct {
		name  string
		state string
		id    client.DataRequestID
	}{
		{"Simulation State", client.SystemStateSim, 1},
		{"Aircraft Loaded", client.SystemStateAircraftLoaded, 2},
		{"Flight Plan", client.SystemStateFlightPlan, 3},
		{"Dialog Mode", client.SystemStateDialogMode, 4},
		{"Flight Loaded", client.SystemStateFlightLoaded, 5},
	}
	for _, test := range testRequests {
		fmt.Printf("Requesting %s...", test.name)
		if err := simclient.RequestSystemState(test.id, test.state); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
		} else {
			fmt.Printf(" ✅ Success\n")
		}
	}
	fmt.Println("\n✅ System state requests sent successfully!")
	fmt.Println("📡 Now attempting to read the responses from SimConnect...")

	fmt.Println("\n=== Reading System State Responses ===")

	// Give SimConnect some time to process requests
	fmt.Println("⏳ Waiting for responses...")
	time.Sleep(500 * time.Millisecond)
	// Try to read responses for a few seconds
	responseCount := 0
	for attempts := 0; attempts < 20; attempts++ {
		response, err := simclient.GetNextDispatch()
		if err != nil {
			fmt.Printf("❌ Error reading response: %v\n", err)
			break
		}
		if response != nil {
			responseCount++
			fmt.Printf("✅ Response %d:\n", responseCount)
			fmt.Printf("   Request ID: %d\n", response.RequestID)
			fmt.Printf("   Data Type: %s\n", response.DataType)

			// Provide meaningful interpretation based on request ID
			switch response.RequestID {
			case 1: // Simulation State
				fmt.Printf("   🎮 Simulation State: ")
				switch response.DataType {
				case "integer":
					if response.IntegerValue == 1 {
						fmt.Printf("ACTIVE (User is controlling aircraft)\n")
					} else {
						fmt.Printf("INACTIVE (User is navigating UI/menus)\n")
					}
					fmt.Printf("      Raw Value: %d\n", response.IntegerValue)
				default:
					fmt.Printf("Unexpected data type for simulation state\n")
				}

			case 2: // Aircraft Loaded
				fmt.Printf("   ✈️  Aircraft Loaded: ")
				switch response.DataType {
				case "string":
					if response.StringValue == "" {
						fmt.Printf("No aircraft loaded\n")
					} else {
						fmt.Printf("YES\n")
						fmt.Printf("      Aircraft File: %s\n", response.StringValue)
					}
				default:
					fmt.Printf("Unexpected data type for aircraft loaded\n")
				}

			case 3: // Flight Plan
				fmt.Printf("   🗺️  Flight Plan: ")
				switch response.DataType {
				case "string":
					if response.StringValue == "" {
						fmt.Printf("No active flight plan\n")
					} else {
						fmt.Printf("ACTIVE\n")
						fmt.Printf("      Flight Plan File: %s\n", response.StringValue)
					}
				case "integer":
					if response.IntegerValue == 1 {
						fmt.Printf("ACTIVE (flight plan loaded)\n")
					} else {
						fmt.Printf("No active flight plan\n")
					}
					fmt.Printf("      Raw Value: %d\n", response.IntegerValue)
				default:
					fmt.Printf("Unexpected data type for flight plan\n")
				}

			case 4: // Dialog Mode
				fmt.Printf("   💬 Dialog Mode: ")
				switch response.DataType {
				case "integer":
					if response.IntegerValue == 1 {
						fmt.Printf("ACTIVE (Dialog box is open)\n")
					} else {
						fmt.Printf("INACTIVE (No dialog boxes open)\n")
					}
					fmt.Printf("      Raw Value: %d\n", response.IntegerValue)
				default:
					fmt.Printf("Unexpected data type for dialog mode\n")
				}

			case 5: // Flight Loaded
				fmt.Printf("   🛫 Flight Loaded: ")
				switch response.DataType {
				case "string":
					if response.StringValue == "" {
						fmt.Printf("No flight loaded\n")
					} else {
						fmt.Printf("YES\n")
						fmt.Printf("      Flight File: %s\n", response.StringValue)
					}
				default:
					fmt.Printf("Unexpected data type for flight loaded\n")
				}

			default:
				fmt.Printf("   📊 Unknown Request ID: ")
				switch response.DataType {
				case "string":
					fmt.Printf("String Value: '%s'\n", response.StringValue)
				case "integer":
					fmt.Printf("Integer Value: %d\n", response.IntegerValue)
				case "float":
					fmt.Printf("Float Value: %.2f\n", response.FloatValue)
				}
			}
			fmt.Println()
		}

		// Small delay between polls
		time.Sleep(100 * time.Millisecond)
	}
	if responseCount == 0 {
		fmt.Println("⚠️  No responses received. This might be normal if:")
		fmt.Println("   - The simulator is not in an active flight state")
		fmt.Println("   - The requested data is not available")
		fmt.Println("   - Responses were processed too quickly")
	} else {
		fmt.Printf("🎉 Successfully received %d system state responses!\n", responseCount)
		fmt.Println("\n📋 Summary of what this means:")
		fmt.Println("   - Your SimConnect wrapper is working correctly")
		fmt.Println("   - Communication with Microsoft Flight Simulator is active")
		fmt.Println("   - System state data is being retrieved successfully")
		fmt.Println("   - The GetNextDispatch() function fix was successful!")
	}
	fmt.Println("\n=== Keeping Connection Open for Inspection ===")
	fmt.Println("✓ Connection is now active and can be seen in SimConnect Inspector")
	fmt.Println("✓ You can now check the SimConnect Inspector in MSFS")
	fmt.Printf("✓ Client '%s' with handle 0x%X should be visible\n", simclient.GetName(), simclient.GetHandle())
	fmt.Println("")

	// Keep connection open for inspection
	for i := 30; i > 0; i-- {
		fmt.Printf("\r⏳ Keeping connection alive... %d seconds remaining (Press Ctrl+C to exit early)", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("\r✓ Inspection period complete                                                     \n")
	fmt.Println("\n=== Testing Connection Close ===")

	// Close the connection
	fmt.Println("Closing SimConnect connection...")
	if err := simclient.Close(); err != nil {
		fmt.Printf("❌ Failed to close connection: %v\n", err)
	} else {
		fmt.Println("✅ Connection closed successfully")
		fmt.Printf("✓ Connection state: %v\n", simclient.IsOpen())
	}

	fmt.Println("\n=== Test Complete ===")
}
