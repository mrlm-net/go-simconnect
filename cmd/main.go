package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mrlm-net/go-simconnect/pkg/simconnect"
)

func main() {
	// Create a new SimConnect client
	client := simconnect.NewClient("Go SimConnect Example")

	fmt.Println("Attempting to connect to SimConnect...")

	// Open connection to SimConnect
	if err := client.Open(); err != nil {
		log.Fatalf("Failed to open SimConnect: %v", err)
	}

	fmt.Printf("Successfully connected to SimConnect as '%s'\n", client.GetName())
	fmt.Printf("Connection handle: 0x%X\n", client.GetHandle())

	// Example: Request system state information
	requestID := simconnect.DataRequestID(1)

	fmt.Println("\nRequesting system states...")

	// Request simulation state
	if err := client.RequestSystemState(requestID, simconnect.SystemStateSim); err != nil {
		log.Printf("Failed to request simulation state: %v", err)
	} else {
		fmt.Println("✓ Simulation state requested")
	}

	// Request aircraft loaded
	requestID++
	if err := client.RequestSystemState(requestID, simconnect.SystemStateAircraftLoaded); err != nil {
		log.Printf("Failed to request aircraft loaded: %v", err)
	} else {
		fmt.Println("✓ Aircraft loaded state requested")
	}

	// Request flight plan
	requestID++
	if err := client.RequestSystemState(requestID, simconnect.SystemStateFlightPlan); err != nil {
		log.Printf("Failed to request flight plan: %v", err)
	} else {
		fmt.Println("✓ Flight plan state requested")
	}

	// Give some time for requests to process
	fmt.Println("\nWaiting for responses...")
	time.Sleep(2 * time.Second)

	// Close the connection
	fmt.Println("\nClosing SimConnect connection...")
	if err := client.Close(); err != nil {
		log.Fatalf("Failed to close SimConnect: %v", err)
	}

	fmt.Println("SimConnect connection closed successfully")
}
