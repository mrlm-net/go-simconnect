package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/mrlm-net/go-simconnect/pkg/simconnect"
)

// SimData represents all the flight simulation data we're collecting
type SimData struct {
	Altitude          float64 // feet
	IndicatedAirspeed float64 // knots
	TrueAirspeed      float64 // knots
	GroundSpeed       float64 // knots
	Latitude          float64 // degrees
	Longitude         float64 // degrees
	HeadingMagnetic   float64 // degrees
	HeadingTrue       float64 // degrees
	BankAngle         float64 // degrees
	PitchAngle        float64 // degrees
	VerticalSpeed     float64 // feet per minute
	EngineRPM         float64 // rpm
	ThrottlePosition  float64 // percent
	GearPosition      float64 // bool (0 or 1)
	FlapsPosition     float64 // percent
}

func main() {
	fmt.Println("=== SimConnect Optimized Multi-Variable Test ===")
	fmt.Println("Testing all confirmed working variables together")
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

	// Create a new SimConnect client
	client := simconnect.NewClientWithDLLPath("Optimized Multi-Variable Test", dllPath)

	if err := client.Open(); err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("✅ Connected to SimConnect!")

	// Define data definition ID and request ID
	defineID := simconnect.DataDefinitionID(1)
	requestID := simconnect.SimObjectDataRequestID(1)

	// Add all confirmed working variables to the data definition
	variables := []struct {
		simVar string
		units  string
		name   string
	}{
		{"Plane Altitude", "feet", "Altitude"},
		{"Airspeed Indicated", "knots", "Indicated Airspeed"},
		{"Airspeed True", "knots", "True Airspeed"},
		{"Ground Velocity", "knots", "Ground Speed"},
		{"Plane Latitude", "degrees", "Latitude"},
		{"Plane Longitude", "degrees", "Longitude"},
		{"Plane Heading Degrees Magnetic", "degrees", "Heading Magnetic"},
		{"Plane Heading Degrees True", "degrees", "Heading True"},
		{"Plane Bank Degrees", "degrees", "Bank Angle"},
		{"Plane Pitch Degrees", "degrees", "Pitch Angle"},
		{"Vertical Speed", "feet per minute", "Vertical Speed"},
		{"General Eng RPM:1", "rpm", "Engine RPM"},
		{"General Eng Throttle Lever Position:1", "percent", "Throttle Position"},
		{"Gear Handle Position", "bool", "Gear Position"},
		{"Flaps Handle Percent", "percent", "Flaps Position"},
	}

	fmt.Printf("📊 Setting up data definition with %d variables...\n", len(variables))

	// Add each variable to the data definition
	for i, variable := range variables {
		fmt.Printf("  %2d. Adding %-30s (%s)...", i+1, variable.name, variable.units)
		if err := client.AddToDataDefinition(defineID, variable.simVar, variable.units, simconnect.SIMCONNECT_DATATYPE_FLOAT64); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			return
		}
		fmt.Println(" ✅")
	}

	fmt.Println("\n🚁 Requesting continuous simulation data...")

	// Request data with periodic updates
	if err := client.RequestDataOnSimObject(requestID, defineID, simconnect.SIMCONNECT_OBJECT_ID_USER, simconnect.SIMCONNECT_PERIOD_SIM_FRAME); err != nil {
		fmt.Printf("❌ Failed to request data: %v\n", err)
		return
	}

	fmt.Println("📡 Data request sent successfully!")
	fmt.Println("\n🔄 Receiving simulation data (Press Ctrl+C to stop)...")
	fmt.Println("=" + "====================================" + "=")

	dataCount := 0
	startTime := time.Now()

	// Main loop to receive and display data
	for {
		data, err := client.GetRawDispatch()
		if err != nil {
			fmt.Printf("❌ Error receiving data: %v\n", err)
			continue
		}

		if data != nil {
			msgType, err := simconnect.ParseMessageType(data)
			if err != nil {
				continue
			}

			if msgType == simconnect.SIMCONNECT_RECV_ID_SIMOBJECT_DATA {
				header, simData, err := simconnect.ParseSimObjectData(data)
				if err != nil {
					fmt.Printf("❌ Error parsing data: %v\n", err)
					continue
				}

				if header.DwRequestID == uint32(requestID) && simData != nil {
					dataCount++
					expectedBytes := len(variables) * 8 // 8 bytes per float64

					fmt.Printf("\n📦 Data Packet #%d (%.1fs elapsed)\n", dataCount, time.Since(startTime).Seconds())
					fmt.Printf("   Raw data size: %d bytes (expected: %d bytes)\n", len(simData), expectedBytes)

					if len(simData) >= expectedBytes {
						// Parse all variables as float64
						simValues := SimData{}
						for i := 0; i < len(variables) && i*8+8 <= len(simData); i++ {
							value := *(*float64)(unsafe.Pointer(&simData[i*8]))

							// Assign to appropriate field based on index
							switch i {
							case 0:
								simValues.Altitude = value
							case 1:
								simValues.IndicatedAirspeed = value
							case 2:
								simValues.TrueAirspeed = value
							case 3:
								simValues.GroundSpeed = value
							case 4:
								simValues.Latitude = value
							case 5:
								simValues.Longitude = value
							case 6:
								simValues.HeadingMagnetic = value
							case 7:
								simValues.HeadingTrue = value
							case 8:
								simValues.BankAngle = value
							case 9:
								simValues.PitchAngle = value
							case 10:
								simValues.VerticalSpeed = value
							case 11:
								simValues.EngineRPM = value
							case 12:
								simValues.ThrottlePosition = value
							case 13:
								simValues.GearPosition = value
							case 14:
								simValues.FlapsPosition = value
							}
						}

						// Display the data in a nice format
						fmt.Printf("   🛩️  AIRCRAFT POSITION:\n")
						fmt.Printf("       Altitude:     %8.1f ft   Lat: %9.4f°   Lon: %9.4f°\n",
							simValues.Altitude, simValues.Latitude, simValues.Longitude)

						fmt.Printf("   💨 AIRSPEED & HEADING:\n")
						fmt.Printf("       IAS: %6.1f kt   TAS: %6.1f kt   GS: %6.1f kt   HDG: %6.1f°M\n",
							simValues.IndicatedAirspeed, simValues.TrueAirspeed, simValues.GroundSpeed, simValues.HeadingMagnetic)

						fmt.Printf("   📐 ATTITUDE & VERTICAL:\n")
						fmt.Printf("       Bank: %6.1f°   Pitch: %6.1f°   VS: %+7.0f fpm\n",
							simValues.BankAngle, simValues.PitchAngle, simValues.VerticalSpeed)

						fmt.Printf("   ⚙️  ENGINE & CONTROLS:\n")
						fmt.Printf("       RPM: %7.0f   Throttle: %5.1f%%   Gear: %s   Flaps: %5.1f%%\n",
							simValues.EngineRPM, simValues.ThrottlePosition,
							map[bool]string{true: "DOWN", false: "UP"}[simValues.GearPosition > 0.5],
							simValues.FlapsPosition)

					} else {
						fmt.Printf("   ⚠️  Incomplete data: got %d bytes, expected %d bytes\n", len(simData), expectedBytes)

						// Show what we did get
						numFloats := len(simData) / 8
						fmt.Printf("   📊 Received %d values:\n", numFloats)
						for i := 0; i < numFloats && i < len(variables); i++ {
							value := *(*float64)(unsafe.Pointer(&simData[i*8]))
							fmt.Printf("       %2d. %-20s: %12.3f %s\n", i+1, variables[i].name, value, variables[i].units)
						}
					}

					// Limit the frequency for readability
					if dataCount >= 50 {
						fmt.Printf("\n🏁 Stopping after %d data packets for readability\n", dataCount)
						fmt.Printf("💡 Data retrieval is working successfully!\n")
						break
					}
				}
			} else if msgType == simconnect.SIMCONNECT_RECV_ID_EXCEPTION {
				fmt.Printf("❌ SimConnect Exception received\n")
			}
		}

		// Small delay to prevent overwhelming output
		time.Sleep(200 * time.Millisecond)
	}

	client.SendDebugMessage(fmt.Sprintf("Optimized multi-variable test completed. Received %d data packets.", dataCount))
	fmt.Println("\n✅ Test completed successfully!")
}
