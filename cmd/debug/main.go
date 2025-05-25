package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mrlm-net/go-simconnect/pkg/simconnect"
)

func main() {
	fmt.Println("=== SimConnect DEBUG Test with Sim Console Logging ===")

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

	// Create a new SimConnect client with the local DLL
	client := simconnect.NewClientWithDLLPath("Go Debug Client", dllPath)
	fmt.Printf("✓ Client created: '%s'\n", client.GetName())

	fmt.Println("\n=== Testing Connection ===")

	// Try to open connection
	fmt.Println("Attempting to connect to SimConnect...")
	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to open SimConnect connection: %v\n", err)
		fmt.Println("\nNote: Make sure Microsoft Flight Simulator 2024 is running.")
		return
	}

	fmt.Printf("✅ Successfully connected to SimConnect!\n")
	fmt.Printf("✓ Connection handle: 0x%X\n", client.GetHandle())
	// Send initial debug message to sim console
	fmt.Println("\n=== Testing Debug Logging ===")
	if err := client.SendDebugMessage("Go SimConnect Debug Client Connected!"); err != nil {
		fmt.Printf("⚠️  Debug message not available (this is ok): %v\n", err)
		fmt.Println("   Continuing with debug test...")
	} else {
		fmt.Println("✅ Debug message sent to Windows debug console")
		fmt.Println("   Check DebugView to see the message")
	}

	fmt.Println("\n=== Testing System State Requests ===")
	// Send debug message about what we're going to do
	// Note: Debug messages might not work in all sim states, so we'll skip errors
	client.SendDebugMessage("Starting system state requests...")

	// Test requesting different system states one by one
	testRequests := []struct {
		name  string
		state string
		id    simconnect.DataRequestID
	}{
		{"Simulation State", simconnect.SystemStateSim, 1},
		{"Aircraft Loaded", simconnect.SystemStateAircraftLoaded, 2},
		{"Flight Plan", simconnect.SystemStateFlightPlan, 3},
		{"Dialog Mode", simconnect.SystemStateDialogMode, 4},
		{"Flight Loaded", simconnect.SystemStateFlightLoaded, 5},
	}

	for _, test := range testRequests {
		fmt.Printf("Requesting %s...", test.name)

		// Log to sim console what we're requesting (ignore errors)
		logMsg := fmt.Sprintf("Requesting %s (ID: %d)", test.name, test.id)
		client.SendDebugMessage(logMsg)

		if err := client.RequestSystemState(test.id, test.state); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			client.SendDebugMessage(fmt.Sprintf("ERROR requesting %s: %v", test.name, err))
		} else {
			fmt.Printf(" ✅ Success\n")
			client.SendDebugMessage(fmt.Sprintf("Successfully requested %s", test.name))
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	client.SendDebugMessage("All system state requests sent. Now polling for responses...")

	fmt.Println("\n=== DEBUG: Reading All Messages ===")
	fmt.Println("⏳ Polling for responses...")
	fmt.Println("   Watch the simulator console for debug messages!")

	// Look for responses
	responseCount := 0
	for attempts := 0; attempts < 100; attempts++ {
		response, err := client.GetNextDispatchDebug()
		if err != nil {
			fmt.Printf("❌ Error reading response: %v\n", err)
			client.SendDebugMessage(fmt.Sprintf("ERROR in GetNextDispatch: %v", err))
			break
		}

		if response != nil {
			responseCount++
			fmt.Printf("🎉 SYSTEM STATE Response %d:\n", responseCount)
			fmt.Printf("   Request ID: %d\n", response.RequestID)
			fmt.Printf("   Data Type: %s\n", response.DataType)

			// Log response to sim console
			logMsg := fmt.Sprintf("Response %d: RequestID=%d, Type=%s", responseCount, response.RequestID, response.DataType)
			client.SendDebugMessage(logMsg)

			switch response.DataType {
			case "string":
				fmt.Printf("   String Value: '%s'\n", response.StringValue)
				client.SendDebugMessage(fmt.Sprintf("  String: '%s'", response.StringValue))
			case "integer":
				fmt.Printf("   Integer Value: %d\n", response.IntegerValue)
				client.SendDebugMessage(fmt.Sprintf("  Integer: %d", response.IntegerValue))
			case "float":
				fmt.Printf("   Float Value: %.2f\n", response.FloatValue)
				client.SendDebugMessage(fmt.Sprintf("  Float: %.2f", response.FloatValue))
			}
			fmt.Println()
		}

		// Update progress in sim console every 20 attempts
		if attempts%20 == 0 {
			client.SendDebugMessage(fmt.Sprintf("Polling attempt %d/100...", attempts))
		}

		// Small delay between polls
		time.Sleep(50 * time.Millisecond)
	}

	client.SendDebugMessage(fmt.Sprintf("Polling complete. Found %d system state responses.", responseCount))

	if responseCount == 0 {
		fmt.Println("\n❓ Debug Summary:")
		fmt.Println("   - Connection was successful")
		fmt.Println("   - Requests were sent successfully")
		fmt.Println("   - We checked for 100 polling attempts")
		fmt.Println("   - Check the simulator console (Ctrl+Shift+Z) for detailed debug info")
		client.SendDebugMessage("No system state responses received. Check console for debug info.")
	} else {
		fmt.Printf("✅ Successfully received %d system state responses!\n", responseCount)
		client.SendDebugMessage(fmt.Sprintf("SUCCESS: Received %d system state responses!", responseCount))
	}

	fmt.Println("\n=== Closing Connection ===")
	client.SendDebugMessage("Closing Go SimConnect Debug Client...")

	// Close the connection
	if err := client.Close(); err != nil {
		fmt.Printf("❌ Failed to close connection: %v\n", err)
	} else {
		fmt.Println("✅ Connection closed successfully")
	}

	fmt.Println("\n=== Debug Test Complete ===")
	fmt.Println("Check the simulator console (Ctrl+Shift+Z or Developer Mode) for detailed debug output!")
}
