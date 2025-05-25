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
	fmt.Println("=== FIXED FLIGHT DASHBOARD ===")
	fmt.Println("Real-time flight simulation data display")
	fmt.Println("ASCII characters for better compatibility")
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
	client := simconnect.NewClientWithDLLPath("Fixed Flight Dashboard", dllPath)

	if err := client.Open(); err != nil {
		fmt.Printf("ERROR: Failed to connect to MSFS: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("SUCCESS: Connected to MSFS 2024!")

	// Create flight data manager
	fdm := simconnect.NewFlightDataManager(client)

	// Add standard flight variables
	fmt.Print("SETUP: Configuring flight data variables...")
	if err := fdm.AddStandardVariables(); err != nil {
		fmt.Printf(" FAILED: %v\n", err)
		return
	}
	fmt.Println(" DONE!")

	// Start real-time data collection
	fmt.Print("START: Beginning real-time data collection...")
	if err := fdm.Start(); err != nil {
		fmt.Printf(" FAILED: %v\n", err)
		return
	}
	fmt.Println(" ACTIVE!")

	fmt.Println("\n*** Flight Dashboard Active (Press Ctrl+C to stop) ***")
	fmt.Println("=====================================================")

	// Display loop
	startTime := time.Now()
	displayCount := 0

	for displayCount < 30 { // Run for 30 iterations (about 30 seconds)
		time.Sleep(1 * time.Second)
		displayCount++

		// Get all current data
		variables := fdm.GetAllVariables()
		dataCount, errorCount, lastUpdate := fdm.GetStats()

		// Clear screen for real-time effect (simplified)
		fmt.Print("\033[H\033[2J")

		// Header
		fmt.Printf("=== MSFS 2024 Fixed Flight Dashboard ===\n")
		fmt.Printf("Runtime: %.0f seconds | Display: #%d | Data Points: %d | Errors: %d\n",
			time.Since(startTime).Seconds(), displayCount, dataCount, errorCount)

		if !lastUpdate.IsZero() {
			fmt.Printf("Last Update: %v ago\n", time.Since(lastUpdate).Truncate(time.Millisecond))
		}
		fmt.Println("")

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

		// Display organized data with updated values
		fmt.Printf("AIRCRAFT POSITION:\n")
		if alt, exists := position["Altitude"]; exists && alt.Updated.After(time.Time{}) {
			fmt.Printf("   Altitude:  %8.0f ft\n", alt.Value)
		} else {
			fmt.Printf("   Altitude:  %8s ft\n", "---")
		}

		if lat, exists := position["Latitude"]; exists && lat.Updated.After(time.Time{}) {
			fmt.Printf("   Latitude:  %8.4f deg\n", lat.Value)
		} else {
			fmt.Printf("   Latitude:  %8s deg\n", "---")
		}

		if lon, exists := position["Longitude"]; exists && lon.Updated.After(time.Time{}) {
			fmt.Printf("   Longitude: %8.4f deg\n", lon.Value)
		} else {
			fmt.Printf("   Longitude: %8s deg\n", "---")
		}

		fmt.Printf("\nSPEED & NAVIGATION:\n")
		if ias, exists := speed["Indicated Airspeed"]; exists && ias.Updated.After(time.Time{}) {
			fmt.Printf("   IAS:       %8.1f knots\n", ias.Value)
		} else {
			fmt.Printf("   IAS:       %8s knots\n", "---")
		}

		if gs, exists := speed["Ground Speed"]; exists && gs.Updated.After(time.Time{}) {
			fmt.Printf("   GS:        %8.1f knots\n", gs.Value)
		} else {
			fmt.Printf("   GS:        %8s knots\n", "---")
		}

		if vs, exists := speed["Vertical Speed"]; exists && vs.Updated.After(time.Time{}) {
			fmt.Printf("   VS:        %8.0f fpm\n", vs.Value)
		} else {
			fmt.Printf("   VS:        %8s fpm\n", "---")
		}

		fmt.Printf("\nATTITUDE & HEADING:\n")
		if hdg, exists := attitude["Heading Magnetic"]; exists && hdg.Updated.After(time.Time{}) {
			fmt.Printf("   Heading:   %8.1f deg\n", hdg.Value)
		} else {
			fmt.Printf("   Heading:   %8s deg\n", "---")
		}

		if bank, exists := attitude["Bank Angle"]; exists && bank.Updated.After(time.Time{}) {
			fmt.Printf("   Bank:      %8.1f deg\n", bank.Value)
		} else {
			fmt.Printf("   Bank:      %8s deg\n", "---")
		}

		if pitch, exists := attitude["Pitch Angle"]; exists && pitch.Updated.After(time.Time{}) {
			fmt.Printf("   Pitch:     %8.1f deg\n", pitch.Value)
		} else {
			fmt.Printf("   Pitch:     %8s deg\n", "---")
		}

		fmt.Printf("\nENGINE & CONTROLS:\n")
		if rpm, exists := engine["Engine RPM"]; exists && rpm.Updated.After(time.Time{}) {
			fmt.Printf("   RPM:       %8.0f rpm\n", rpm.Value)
		} else {
			fmt.Printf("   RPM:       %8s rpm\n", "---")
		}

		if thr, exists := engine["Throttle Position"]; exists && thr.Updated.After(time.Time{}) {
			fmt.Printf("   Throttle:  %8.1f %%\n", thr.Value)
		} else {
			fmt.Printf("   Throttle:  %8s %%\n", "---")
		}

		if gear, exists := controls["Gear Position"]; exists && gear.Updated.After(time.Time{}) {
			gearStatus := "UP"
			if gear.Value > 0.5 {
				gearStatus = "DOWN"
			}
			fmt.Printf("   Gear:      %8s\n", gearStatus)
		} else {
			fmt.Printf("   Gear:      %8s\n", "---")
		}

		fmt.Printf("\n=====================================================\n")
		fmt.Printf("Data collection rate: %.1f Hz | Updates/var: %.1f Hz\n",
			float64(dataCount)/time.Since(startTime).Seconds(),
			float64(dataCount)/time.Since(startTime).Seconds()/float64(len(variables)))
	}

	fdm.Stop()

	finalDataCount, finalErrorCount, _ := fdm.GetStats()
	fmt.Printf("\nFINAL DASHBOARD STATISTICS:\n")
	fmt.Printf("  Total data points collected: %d\n", finalDataCount)
	fmt.Printf("  Total errors: %d\n", finalErrorCount)
	fmt.Printf("  Runtime: %.1f seconds\n", time.Since(startTime).Seconds())
	fmt.Printf("  Average data rate: %.1f Hz\n", float64(finalDataCount)/time.Since(startTime).Seconds())

	fmt.Println("\nFixed Flight Dashboard completed successfully!")

	client.SendDebugMessage(fmt.Sprintf("Fixed dashboard finished. Collected %d data points with %d errors.",
		finalDataCount, finalErrorCount))
}
