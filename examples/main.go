package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mrlm-net/go-simconnect/pkg/client"
)

func main() { // Create a new SimConnect client
	simclient := client.NewClientWithDLLPath("Go SimConnect Example", "C:\\MSFS 2024 SDK\\SimConnect SDK\\lib\\SimConnect.dll")

	fmt.Println("Attempting to connect to SimConnect...")
	// Open connection to SimConnect
	if err := simclient.Open(); err != nil {
		log.Fatalf("Failed to open SimConnect: %v", err)
	}

	fmt.Printf("Successfully connected to SimConnect as '%s'\n", simclient.GetName())
	fmt.Printf("Connection handle: 0x%X\n", simclient.GetHandle())
	// Example: Request system state information
	requestID := client.DataRequestID(1)

	fmt.Println("\nRequesting system states...")

	// Request simulation state
	if err := simclient.RequestSystemState(requestID, client.SystemStateSim); err != nil {
		log.Printf("Failed to request simulation state: %v", err)
	} else {
		fmt.Println("✓ Simulation state requested")
	}

	// Request aircraft loaded
	requestID++
	if err := simclient.RequestSystemState(requestID, client.SystemStateAircraftLoaded); err != nil {
		log.Printf("Failed to request aircraft loaded: %v", err)
	} else {
		fmt.Println("✓ Aircraft loaded state requested")
	}

	// Request flight plan
	requestID++
	if err := simclient.RequestSystemState(requestID, client.SystemStateFlightPlan); err != nil {
		log.Printf("Failed to request flight plan: %v", err)
	} else {
		fmt.Println("✓ Flight plan state requested")
	}

	// Give some time for requests to process
	fmt.Println("\nWaiting for responses...")
	time.Sleep(2 * time.Second)
	// Close the connection
	fmt.Println("\nClosing SimConnect connection...")
	if err := simclient.Close(); err != nil {
		log.Fatalf("Failed to close SimConnect: %v", err)
	}

	fmt.Println("SimConnect connection closed successfully")
}
