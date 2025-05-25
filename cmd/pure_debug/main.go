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
	fmt.Println("=== SimConnect PURE DEBUG Test ===")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	dllPath := filepath.Join(wd, "lib", "SimConnect.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		log.Fatalf("SimConnect.dll not found at %s", dllPath)
	}

	client := simconnect.NewClientWithDLLPath("Go Pure Debug", dllPath)

	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("✅ Connected! Handle: 0x%X\n", client.GetHandle())

	// Send one simple request
	fmt.Print("Requesting Simulation State...")
	if err := client.RequestSystemState(1, simconnect.SystemStateSim); err != nil {
		fmt.Printf(" ❌ Failed: %v\n", err)
		return
	}
	fmt.Println(" ✅ Success")

	// Poll for responses
	fmt.Println("🔍 Polling for responses (20 attempts)...")
	responseCount := 0

	for i := 0; i < 20; i++ {
		response, err := client.GetNextDispatchDebug()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			break
		}

		if response != nil {
			responseCount++
			fmt.Printf("✅ Response %d: ID=%d, Type=%s\n", responseCount, response.RequestID, response.DataType)
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("📊 Total responses: %d\n", responseCount)
}
