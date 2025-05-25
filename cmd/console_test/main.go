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
	fmt.Println("IMPORTANT: SimConnect does not have a built-in function to send messages to the MSFS console.")
	fmt.Println("This test demonstrates debug logging using Windows OutputDebugString instead.")
	fmt.Println("Debug messages can be viewed with tools like DebugView or Visual Studio.")
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
	client := simconnect.NewClientWithDLLPath("Debug Test Client", dllPath)
	fmt.Printf("✓ Client created: '%s'\n", client.GetName())

	fmt.Println("\n=== Testing Connection ===")

	// Try to open connection
	fmt.Println("Attempting to connect to SimConnect...")
	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to open SimConnect connection: %v\n", err)
		fmt.Println("\nNote: This is expected if Microsoft Flight Simulator is not running.")
		fmt.Println("To test debug logging:")
		fmt.Println("1. Start Microsoft Flight Simulator 2024")
		fmt.Println("2. Download and run DebugView from Microsoft Sysinternals")
		fmt.Println("3. Run this test again")
		fmt.Println("4. Check DebugView for debug messages")
		return
	}
	fmt.Printf("✅ Successfully connected to SimConnect!\n")
	fmt.Printf("✓ Connection handle: 0x%X\n", client.GetHandle())

	fmt.Println("\n=== Testing Debug Message Logging ===")
	fmt.Println("📝 Now sending debug messages using Windows OutputDebugString...")
	fmt.Println("   💡 To see messages: Download and run DebugView from Microsoft Sysinternals")
	fmt.Println("   🔗 https://docs.microsoft.com/en-us/sysinternals/downloads/debugview")
	fmt.Println("")

	// Test messages to send to the debug console
	testMessages := []string{
		"Hello from Go SimConnect Client!",
		"Debug logging test started",
		"Testing message with special characters: Test123",
		"SimConnect wrapper is working correctly",
		"Flight simulation data can be accessed",
		"All systems operational!",
	}
	// Send each test message with a delay
	for i, message := range testMessages {
		fmt.Printf("Sending debug message %d: %s\n", i+1, message)

		if err := client.SendDebugMessage(message); err != nil {
			fmt.Printf("❌ Failed to send debug message: %v\n", err)
		} else {
			fmt.Printf("✅ Debug message sent successfully\n")
		}

		// Wait a bit between messages for easier reading in debug output
		time.Sleep(1 * time.Second)
		fmt.Println("")
	}
	fmt.Println("🎯 Test Summary:")
	fmt.Printf("   - Sent %d debug messages using OutputDebugString\n", len(testMessages))
	fmt.Println("   - Check DebugView or Visual Studio Output window to see messages")
	fmt.Println("   - Messages should appear with [SimConnect:Debug Test Client] prefix")
	fmt.Println("")

	fmt.Println("=== Final Verification Test ===")
	finalMessage := fmt.Sprintf("Debug test completed at %s", time.Now().Format("15:04:05"))
	fmt.Printf("Sending final debug message: %s\n", finalMessage)

	if err := client.SendDebugMessage(finalMessage); err != nil {
		fmt.Printf("❌ Failed to send final debug message: %v\n", err)
	} else {
		fmt.Printf("✅ Final debug message sent successfully\n")
	}

	fmt.Println("\n=== Keeping Connection Open for Debug Inspection ===")
	fmt.Println("✓ Connection is active - check DebugView now!")
	fmt.Println("📋 Instructions:")
	fmt.Println("   1. Download DebugView from https://docs.microsoft.com/en-us/sysinternals/downloads/debugview")
	fmt.Println("   2. Run DebugView as Administrator")
	fmt.Println("   3. In DebugView, enable 'Capture Win32' (Ctrl+W)")
	fmt.Println("   4. Look for messages with [SimConnect:Debug Test Client] prefix")
	fmt.Println("   5. You should see all the debug messages that were just sent")
	fmt.Println("")
	// Keep connection open for inspection
	for i := 10; i > 0; i-- {
		fmt.Printf("\r⏳ Keeping connection alive for debug inspection... %d seconds remaining", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("\r✓ Debug inspection period complete                                        \n")

	fmt.Println("\n=== Testing Connection Close ===")

	// Send a goodbye message
	goodbyeMessage := "Go SimConnect Client disconnecting..."
	fmt.Printf("Sending goodbye debug message: %s\n", goodbyeMessage)
	if err := client.SendDebugMessage(goodbyeMessage); err != nil {
		fmt.Printf("❌ Failed to send goodbye debug message: %v\n", err)
	} else {
		fmt.Printf("✅ Goodbye debug message sent\n")
	}

	// Close the connection
	fmt.Println("Closing SimConnect connection...")
	if err := client.Close(); err != nil {
		fmt.Printf("❌ Failed to close connection: %v\n", err)
	} else {
		fmt.Println("✅ Connection closed successfully")
	}
	fmt.Println("\n=== Debug Message Test Complete ===")
	fmt.Println("🎉 If you saw messages in DebugView, the debug logging function works perfectly!")
	fmt.Println("")
	fmt.Println("📚 Summary of findings:")
	fmt.Println("   • SimConnect does NOT have a built-in console logging function")
	fmt.Println("   • Use Windows OutputDebugString + DebugView for debug logging")
	fmt.Println("   • Alternative: Write to log files or use standard output")
}
