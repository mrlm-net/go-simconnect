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
	fmt.Println("=== FIXED FLIGHT DATA MANAGER TEST ===")
	fmt.Println("Testing the corrected pointer bug fix")
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
	client := simconnect.NewClientWithDLLPath("Fixed Manager Test", dllPath)

	if err := client.Open(); err != nil {
		fmt.Printf("FAILED to connect to MSFS: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("SUCCESS: Connected to MSFS 2024!")

	// Create flight data manager
	fdm := simconnect.NewFlightDataManager(client)

	// Add just a few critical variables for testing
	fmt.Print("Adding test variables...")
	if err := fdm.AddVariable("Altitude", "Plane Altitude", "feet"); err != nil {
		log.Fatalf("Failed to add Altitude: %v", err)
	}
	if err := fdm.AddVariable("Latitude", "Plane Latitude", "degrees"); err != nil {
		log.Fatalf("Failed to add Latitude: %v", err)
	}
	if err := fdm.AddVariable("Longitude", "Plane Longitude", "degrees"); err != nil {
		log.Fatalf("Failed to add Longitude: %v", err)
	}
	fmt.Println(" Done!")

	// Start data collection
	fmt.Print("Starting data collection...")
	if err := fdm.Start(); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	fmt.Println(" Started!")

	fmt.Println("\nCollecting data for 10 seconds...")
	fmt.Println("=====================================")

	// Monitor for 10 seconds
	startTime := time.Now()
	for time.Since(startTime) < 10*time.Second {
		time.Sleep(1 * time.Second)

		variables := fdm.GetAllVariables()
		dataCount, errorCount, lastUpdate := fdm.GetStats()

		fmt.Printf("\nData Update (%.0fs elapsed):\n", time.Since(startTime).Seconds())
		fmt.Printf("  Stats: %d total updates, %d errors\n", dataCount, errorCount)
		if !lastUpdate.IsZero() {
			fmt.Printf("  Last update: %v ago\n", time.Since(lastUpdate).Truncate(time.Millisecond))
		}

		for _, variable := range variables {
			if variable.Updated.After(time.Time{}) {
				fmt.Printf("  %-15s: %12.3f %s (updated %v ago)\n",
					variable.Name, variable.Value, variable.Units,
					time.Since(variable.Updated).Truncate(time.Millisecond))
			} else {
				fmt.Printf("  %-15s: %12s %s (no data yet)\n",
					variable.Name, "---", variable.Units)
			}
		}
	}

	fdm.Stop()

	finalDataCount, finalErrorCount, _ := fdm.GetStats()
	fmt.Printf("\nFINAL RESULTS:\n")
	fmt.Printf("  Total data points: %d\n", finalDataCount)
	fmt.Printf("  Total errors: %d\n", finalErrorCount)

	if finalDataCount > 0 {
		fmt.Println("  SUCCESS: FlightDataManager is working correctly!")
	} else {
		fmt.Println("  PROBLEM: No data collected - check MSFS connection")
	}
}
