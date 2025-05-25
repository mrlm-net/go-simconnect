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
	fmt.Println("=== MSFS 2024 Flight Dashboard ===")
	fmt.Println("Real-time flight simulation data display")
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

	// Create SimConnect client
	client := simconnect.NewClientWithDLLPath("Flight Dashboard", dllPath)

	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to connect to MSFS: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✅ Connected to MSFS 2024!")

	// Create flight data manager
	fdm := simconnect.NewFlightDataManager(client)

	// Add standard flight variables
	fmt.Print("📊 Setting up flight data variables...")
	if err := fdm.AddStandardVariables(); err != nil {
		fmt.Printf(" ❌ Failed: %v\n", err)
		return
	}
	fmt.Println(" ✅ Done!")

	// Start real-time data collection
	fmt.Print("🚁 Starting real-time data collection...")
	if err := fdm.Start(); err != nil {
		fmt.Printf(" ❌ Failed: %v\n", err)
		return
	}
	fmt.Println(" ✅ Started!")

	fmt.Println("\n🎮 Flight Dashboard Active (Press Ctrl+C to stop)")
	fmt.Println("=" + "================================" + "=")

	// Display loop
	startTime := time.Now()
	for i := 0; i < 60; i++ { // Run for 60 iterations (about 1 minute)
		time.Sleep(1 * time.Second)

		// Clear screen for real-time effect (simplified for demo)
		fmt.Print("\033[H\033[2J")

		// Header
		fmt.Printf("=== MSFS 2024 Flight Dashboard ===\n")
		fmt.Printf("Runtime: %.0f seconds | Updates: %d\n\n", time.Since(startTime).Seconds(), i+1)

		// Get all current data
		variables := fdm.GetAllVariables()

		// Group data by category
		position := make(map[string]simconnect.FlightVariable)
		speed := make(map[string]simconnect.FlightVariable)
		attitude := make(map[string]simconnect.FlightVariable)
		engine := make(map[string]simconnect.FlightVariable)
		controls := make(map[string]simconnect.FlightVariable)

		for _, variable := range variables {
			switch variable.Name {
			case "Altitude", "Latitude", "Longitude":
				position[variable.Name] = variable
			case "Indicated Airspeed", "True Airspeed", "Ground Speed", "Vertical Speed":
				speed[variable.Name] = variable
			case "Heading Magnetic", "Heading True", "Bank Angle", "Pitch Angle":
				attitude[variable.Name] = variable
			case "Engine RPM", "Throttle Position":
				engine[variable.Name] = variable
			case "Gear Position", "Flaps Position":
				controls[variable.Name] = variable
			}
		}

		// Display organized data
		fmt.Printf("🛩️  AIRCRAFT POSITION\n")
		fmt.Printf("   Altitude:  %8.0f ft\n", position["Altitude"].Value)
		fmt.Printf("   Latitude:  %9.4f°\n", position["Latitude"].Value)
		fmt.Printf("   Longitude: %9.4f°\n", position["Longitude"].Value)
		fmt.Println()

		fmt.Printf("💨 SPEED & NAVIGATION\n")
		fmt.Printf("   Indicated Airspeed: %6.0f knots\n", speed["Indicated Airspeed"].Value)
		fmt.Printf("   True Airspeed:      %6.0f knots\n", speed["True Airspeed"].Value)
		fmt.Printf("   Ground Speed:       %6.0f knots\n", speed["Ground Speed"].Value)
		fmt.Printf("   Vertical Speed:     %+6.0f fpm\n", speed["Vertical Speed"].Value)
		fmt.Printf("   Heading (Magnetic): %6.1f°\n", attitude["Heading Magnetic"].Value)
		fmt.Printf("   Heading (True):     %6.1f°\n", attitude["Heading True"].Value)
		fmt.Println()

		fmt.Printf("📐 AIRCRAFT ATTITUDE\n")
		fmt.Printf("   Bank Angle:  %+6.1f°\n", attitude["Bank Angle"].Value)
		fmt.Printf("   Pitch Angle: %+6.1f°\n", attitude["Pitch Angle"].Value)
		fmt.Println()

		fmt.Printf("⚙️  ENGINE & CONTROLS\n")
		fmt.Printf("   Engine RPM:       %7.0f rpm\n", engine["Engine RPM"].Value)
		fmt.Printf("   Throttle:         %6.1f%%\n", engine["Throttle Position"].Value)
		gearStatus := "UP"
		if controls["Gear Position"].Value > 0.5 {
			gearStatus = "DOWN"
		}
		fmt.Printf("   Landing Gear:     %s\n", gearStatus)
		fmt.Printf("   Flaps:            %6.1f%%\n", controls["Flaps Position"].Value)
		fmt.Println()

		// Data freshness indicators
		fmt.Printf("📡 DATA STATUS\n")
		allFresh := true
		for _, variable := range variables {
			age := time.Since(variable.Updated)
			status := "✅"
			if age > 5*time.Second {
				status = "⚠️ "
				allFresh = false
			} else if age > 10*time.Second {
				status = "❌"
				allFresh = false
			}

			if !allFresh && status != "✅" {
				fmt.Printf("   %s %-20s (%.1fs ago)\n", status, variable.Name, age.Seconds())
			}
		}

		if allFresh {
			fmt.Printf("   ✅ All data current (last update < 5s ago)\n")
		}

		fmt.Printf("\n💡 Real-time flight simulation data from MSFS 2024\n")
		fmt.Printf("🔄 Data automatically refreshes every second\n")
	}

	// Stop data collection
	fmt.Print("\n🛑 Stopping data collection...")
	fdm.Stop()
	fmt.Println(" ✅ Stopped!")

	client.SendDebugMessage("Flight Dashboard completed successfully.")
	fmt.Println("\n✅ Flight Dashboard session completed!")
}
