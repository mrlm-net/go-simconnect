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
	fmt.Println("=== SimConnect Debug Message Test ===")
	fmt.Println("Testing debug message functionality with different string formats")
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
	client := simconnect.NewClientWithDLLPath("String Test Client", dllPath)
	fmt.Printf("✓ Client created: '%s'\n", client.GetName())

	fmt.Println("\n=== Testing Connection ===")

	// Try to open connection
	fmt.Println("Attempting to connect to SimConnect...")
	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to open SimConnect connection: %v\n", err)
		fmt.Println("\nNote: This is expected if Microsoft Flight Simulator is not running.")
		return
	}

	fmt.Printf("✅ Successfully connected to SimConnect!\n")
	fmt.Printf("✓ Connection handle: 0x%X\n", client.GetHandle())

	fmt.Println("\n=== Testing Simple ASCII Strings ===")

	// Test with very simple ASCII strings first
	simpleMessages := []string{
		"Hello",
		"Test message",
		"SimConnect debug test",
		"ASCII only message 123",
	}

	for i, message := range simpleMessages {
		fmt.Printf("Test %d: Sending '%s'\n", i+1, message)

		if err := client.SendDebugMessage(message); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Success!\n")
		}

		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\n=== Testing Empty and Special Cases ===")

	specialCases := []string{
		"",              // Empty string
		" ",             // Single space
		"A",             // Single character
		"123",           // Numbers only
		"Test\nNewline", // With newline
		"Test\tTab",     // With tab
	}

	for i, message := range specialCases {
		displayMsg := message
		if displayMsg == "" {
			displayMsg = "(empty string)"
		}
		displayMsg = fmt.Sprintf("%q", displayMsg) // Show escaped version

		fmt.Printf("Special test %d: %s\n", i+1, displayMsg)

		if err := client.SendDebugMessage(message); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Success!\n")
		}

		time.Sleep(500 * time.Millisecond)
	}
	// Try to send a success message if any worked
	fmt.Println("\n=== Final Test ===")
	finalMessage := "Debug message test completed"
	fmt.Printf("Sending final message: %s\n", finalMessage)

	if err := client.SendDebugMessage(finalMessage); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else {
		fmt.Printf("✅ Success!\n")
	}

	fmt.Println("\n=== Instructions for Verification ===")
	fmt.Println("📋 To check if any messages appeared in debug output:")
	fmt.Println("   1. Download and run DebugView from Microsoft Sysinternals")
	fmt.Println("   2. Enable 'Capture Win32' in DebugView")
	fmt.Println("   3. Look for messages with [SimConnect:] prefix")
	fmt.Println("   4. If you see messages, debug logging is working!")
	fmt.Println("")

	// Keep connection open briefly for inspection
	for i := 10; i > 0; i-- {
		fmt.Printf("\r⏳ Keeping connection alive for verification... %d seconds", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("\r✓ Test complete                                            \n")

	// Close the connection
	fmt.Println("\nClosing connection...")
	if err := client.Close(); err != nil {
		fmt.Printf("❌ Failed to close: %v\n", err)
	} else {
		fmt.Println("✅ Connection closed successfully")
	}

	fmt.Println("\n=== String Test Complete ===")
}
