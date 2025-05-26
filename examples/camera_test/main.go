package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() {
	fmt.Println("=== 📹 CAMERA CONTROL TEST ===")
	fmt.Println("This test will cycle through different camera views")
	fmt.Println("👀 Watch for IMMEDIATE camera changes in MSFS 2024!")
	fmt.Println("🎯 This is the DEFINITIVE test for SetData functionality")
	fmt.Println()

	// Create SimConnect client with MSFS 2024 SDK path
	dllPath := `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
	simClient := client.NewClientWithDLLPath("CameraTest", dllPath)

	// Connect to SimConnect
	fmt.Println("🔗 Connecting to SimConnect...")
	if err := simClient.Open(); err != nil {
		log.Fatalf("❌ Failed to connect to SimConnect: %v", err)
	}
	defer simClient.Close()
	fmt.Println("✅ Connected successfully")

	// Create flight data manager
	fdm := client.NewFlightDataManager(simClient)

	// Add camera state variable (writable)
	fmt.Println("📝 Adding Camera State variable...")
	if err := fdm.AddVariableWithWritable("Camera State", "Camera State", "number", true); err != nil {
		log.Fatalf("❌ Failed to add camera state variable: %v", err)
	}
	fmt.Println("✅ Camera State variable added")

	// Start data collection
	fmt.Println("🚀 Starting data monitoring...")
	if err := fdm.Start(); err != nil {
		log.Fatalf("❌ Failed to start data collection: %v", err)
	}
	defer fdm.Stop()

	// Wait for initial data
	fmt.Println("⏱️ Waiting for initial data...")
	time.Sleep(3 * time.Second)

	// Check current camera state
	cameraVar, found := fdm.GetVariable("Camera State")
	if !found {
		log.Fatal("❌ Camera State variable not found")
	}

	fmt.Printf("📊 Current Camera State: %.0f\n", cameraVar.Value)
	fmt.Println()

	// Camera test sequence with 15 second pauses
	cameraSequence := []struct {
		state       float64
		description string
		viewName    string
	}{
		{2.0, "🏠 Switching to COCKPIT view", "Cockpit"},
		{3.0, "🌍 Switching to EXTERNAL view", "External"},
		{4.0, "🪶 Switching to WING view", "Wing"},
		{5.0, "🚁 Switching to TAIL view", "Tail"},
		{6.0, "🗼 Switching to TOWER view", "Tower"},
		{2.0, "🏠 Returning to COCKPIT view", "Cockpit"},
	}

	fmt.Println("🎬 Starting Camera Control Test Sequence!")
	fmt.Println("⏰ Each camera change will last 15 seconds")
	fmt.Println("👁️ Watch your simulator screen for IMMEDIATE camera changes!")
	fmt.Println()

	for i, step := range cameraSequence {
		fmt.Printf("📹 Step %d: %s (State %.0f)\n", i+1, step.description, step.state)
		fmt.Printf("   🎯 Target: %s View\n", step.viewName)

		// Send the camera change command
		fmt.Printf("   🎛️  Setting camera state to: %.0f\n", step.state)
		if err := fdm.SetVariable("Camera State", step.state); err != nil {
			fmt.Printf("   ❌ ERROR: %v\n", err)
			continue
		}
		fmt.Printf("   ✅ SetData command sent successfully\n")

		// Brief pause to let the command take effect
		time.Sleep(1 * time.Second)

		// Read back the camera state to verify
		cameraVar, found := fdm.GetVariable("Camera State")
		if found {
			fmt.Printf("   📖 Camera State readback: %.0f\n", cameraVar.Value)

			if cameraVar.Value == step.state {
				fmt.Printf("   🎯 SUCCESS: Camera state changed to %.0f (%s) as expected\n", step.state, step.viewName)
			} else {
				fmt.Printf("   ⚠️  Note: Expected %.0f but got %.0f (may take time to update)\n", step.state, cameraVar.Value)
			}
		} else {
			fmt.Printf("   ❌ ERROR: Could not read camera state\n")
		}

		fmt.Printf("   👀 LOOK AT YOUR SIMULATOR NOW - Should be in %s view!\n", step.viewName)
		fmt.Printf("   ⏰ Observing for 15 seconds...\n")
		fmt.Println()

		// Wait 15 seconds for visual observation
		for countdown := 15; countdown > 0; countdown-- {
			if countdown%5 == 0 || countdown <= 3 {
				fmt.Printf("   ⏳ %d seconds remaining in %s view...\n", countdown, step.viewName)
			}
			time.Sleep(1 * time.Second)
		}

		fmt.Println("   ✅ Observation period complete")
		fmt.Println()
	}

	fmt.Println("🏁 Camera Control Test Sequence COMPLETED!")
	fmt.Println()

	// Final verification
	cameraVar, found = fdm.GetVariable("Camera State")
	if found {
		var viewName string
		switch cameraVar.Value {
		case 2:
			viewName = "Cockpit"
		case 3:
			viewName = "External"
		case 4:
			viewName = "Wing"
		case 5:
			viewName = "Tail"
		case 6:
			viewName = "Tower"
		default:
			viewName = fmt.Sprintf("Unknown (%.0f)", cameraVar.Value)
		}
		fmt.Printf("🎯 Final Camera State: %.0f (%s)\n", cameraVar.Value, viewName)
	}
	fmt.Println()

	// Results evaluation
	fmt.Println("📊 TEST RESULTS EVALUATION:")
	fmt.Println("✅ If you saw the camera views changing every 15 seconds:")
	fmt.Println("   → SetData is WORKING PERFECTLY! 🎉")
	fmt.Println("   → Commands are reaching the simulator!")
	fmt.Println("   → go-simconnect SetData implementation is VALIDATED!")
	fmt.Println()
	fmt.Println("❌ If camera views did NOT change:")
	fmt.Println("   → SetData may not be working properly")
	fmt.Println("   → Need to investigate further")
	fmt.Println()
	fmt.Println("🔄 Continuous monitoring (Press Ctrl+C to exit):")
	fmt.Println("Monitoring camera state for any changes...")

	// Continue monitoring camera state
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cameraVar, found := fdm.GetVariable("Camera State")
		if found {
			var viewName string
			switch cameraVar.Value {
			case 2:
				viewName = "Cockpit"
			case 3:
				viewName = "External"
			case 4:
				viewName = "Wing"
			case 5:
				viewName = "Tail"
			case 6:
				viewName = "Tower"
			default:
				viewName = fmt.Sprintf("Unknown (%.0f)", cameraVar.Value)
			}

			age := time.Since(cameraVar.Updated)
			fmt.Printf("[%s] Camera: %.0f (%s) - Updated %v ago\n",
				time.Now().Format("15:04:05"),
				cameraVar.Value,
				viewName,
				age.Truncate(time.Second))
		}
	}
}
