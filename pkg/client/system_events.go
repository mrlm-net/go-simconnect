// filepath: pkg/client/system_events.go
package client

import (
	"fmt"
	"sync"
	"time"
)

// SystemEventManager provides thread-safe management of SimConnect system events
type SystemEventManager struct {
	client     *Client                                            // SimConnect client
	mutex      sync.RWMutex                                       // Thread safety
	callbacks  map[SIMCONNECT_CLIENT_EVENT_ID]SystemEventCallback // Event callbacks
	eventNames map[SIMCONNECT_CLIENT_EVENT_ID]string              // Event ID to name mapping
	running    bool                                               // Manager state
	stopChan   chan struct{}                                      // Stop signal
	errorChan  chan error                                         // Error notifications
	nextID     SIMCONNECT_CLIENT_EVENT_ID                         // Next available event ID
}

// NewSystemEventManager creates a new SystemEventManager instance
func NewSystemEventManager(client *Client) *SystemEventManager {
	return &SystemEventManager{
		client:     client,
		callbacks:  make(map[SIMCONNECT_CLIENT_EVENT_ID]SystemEventCallback),
		eventNames: make(map[SIMCONNECT_CLIENT_EVENT_ID]string),
		running:    false,
		stopChan:   make(chan struct{}),
		errorChan:  make(chan error, 10), // Buffered channel for non-blocking errors
		nextID:     1000,                 // Start at 1000 to avoid conflicts
	}
}

// SubscribeToEvent subscribes to a system event with a callback
func (sem *SystemEventManager) SubscribeToEvent(eventName string, callback SystemEventCallback) (SIMCONNECT_CLIENT_EVENT_ID, error) {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	if !sem.client.IsOpen() {
		return 0, fmt.Errorf("SimConnect client is not open")
	}

	// Assign new event ID
	eventID := sem.nextID
	sem.nextID++

	// Subscribe to the event via SimConnect
	if err := sem.client.SubscribeToSystemEvent(eventID, eventName); err != nil {
		return 0, fmt.Errorf("failed to subscribe to event '%s': %v", eventName, err)
	}

	// Store callback and name mapping
	sem.callbacks[eventID] = callback
	sem.eventNames[eventID] = eventName

	return eventID, nil
}

// UnsubscribeFromEvent unsubscribes from a system event
func (sem *SystemEventManager) UnsubscribeFromEvent(eventID SIMCONNECT_CLIENT_EVENT_ID) error {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	if !sem.client.IsOpen() {
		return fmt.Errorf("SimConnect client is not open")
	}

	// Unsubscribe from the event via SimConnect
	if err := sem.client.UnsubscribeFromSystemEvent(eventID); err != nil {
		return fmt.Errorf("failed to unsubscribe from event ID %d: %v", eventID, err)
	}

	// Remove from internal tracking
	delete(sem.callbacks, eventID)
	delete(sem.eventNames, eventID)

	return nil
}

// SetEventState sets the state of a system event (ON/OFF)
func (sem *SystemEventManager) SetEventState(eventID SIMCONNECT_CLIENT_EVENT_ID, state SIMCONNECT_STATE) error {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()

	if !sem.client.IsOpen() {
		return fmt.Errorf("SimConnect client is not open")
	}

	// Check if event is known
	if _, exists := sem.callbacks[eventID]; !exists {
		return fmt.Errorf("event ID %d is not subscribed", eventID)
	}

	return sem.client.SetSystemEventState(eventID, state)
}

// Start begins processing system events in a background goroutine
func (sem *SystemEventManager) Start() error {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	if sem.running {
		return fmt.Errorf("SystemEventManager is already running")
	}

	if !sem.client.IsOpen() {
		return fmt.Errorf("SimConnect client is not open")
	}

	sem.running = true
	sem.stopChan = make(chan struct{})

	// Start event processing goroutine
	go sem.eventLoop()

	return nil
}

// Stop halts the system event processing
func (sem *SystemEventManager) Stop() {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	if !sem.running {
		return
	}

	sem.running = false
	close(sem.stopChan)
}

// IsRunning returns whether the event manager is currently running
func (sem *SystemEventManager) IsRunning() bool {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()
	return sem.running
}

// GetErrors returns the error channel for monitoring runtime errors
func (sem *SystemEventManager) GetErrors() <-chan error {
	return sem.errorChan
}

// GetSubscribedEvents returns a copy of currently subscribed events
func (sem *SystemEventManager) GetSubscribedEvents() map[SIMCONNECT_CLIENT_EVENT_ID]string {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()

	// Return a copy to prevent external modification
	events := make(map[SIMCONNECT_CLIENT_EVENT_ID]string)
	for id, name := range sem.eventNames {
		events[id] = name
	}
	return events
}

// eventLoop is the main event processing loop that runs in a background goroutine
// This integrates with the existing dispatch mechanism to avoid competing message polling
func (sem *SystemEventManager) eventLoop() {
	ticker := time.NewTicker(50 * time.Millisecond) // Poll at 20Hz
	defer ticker.Stop()

	for {
		select {
		case <-sem.stopChan:
			return

		case <-ticker.C:
			// Process all available events in this tick
			if err := sem.processEventsFromRawDispatch(); err != nil {
				// Send error to error channel (non-blocking)
				select {
				case sem.errorChan <- err:
				default:
					// Channel full, skip this error
				}
			}
		}
	}
}

// processEventsFromRawDispatch processes system events directly from raw dispatch data
// This integrates with the existing FlightDataManager dispatch mechanism to avoid message conflicts
func (sem *SystemEventManager) processEventsFromRawDispatch() error {
	for {
		// Get raw dispatch data directly (same as FlightDataManager)
		data, err := sem.client.GetRawDispatch()
		if err != nil {
			return fmt.Errorf("error getting raw dispatch: %v", err)
		}

		if data == nil {
			// No more messages available
			break
		}

		// Parse message type
		msgType, err := ParseMessageType(data)
		if err != nil {
			return fmt.Errorf("error parsing message type: %v", err)
		}

		// Only process event messages
		switch msgType {
		case SIMCONNECT_RECV_ID_EVENT,
			SIMCONNECT_RECV_ID_EVENT_FILENAME,
			SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE,
			SIMCONNECT_RECV_ID_EVENT_FRAME:

			// Parse the event using existing GetSystemEvent logic
			eventData, err := sem.parseEventFromRawData(data, msgType)
			if err != nil {
				return fmt.Errorf("error parsing event data: %v", err)
			}

			if eventData != nil {
				// Find and execute callback
				sem.mutex.RLock()
				callback, exists := sem.callbacks[eventData.EventID]
				eventName, nameExists := sem.eventNames[eventData.EventID]
				sem.mutex.RUnlock()

				if exists && callback != nil {
					// Update event data with human-readable name
					if nameExists {
						eventData.EventName = eventName
					}

					// Execute callback in a separate goroutine to prevent blocking
					go func(event SystemEventData, cb SystemEventCallback) {
						defer func() {
							if r := recover(); r != nil {
								// Send panic as error to error channel
								select {
								case sem.errorChan <- fmt.Errorf("event callback panic: %v", r):
								default:
								}
							}
						}()
						cb(event)
					}(*eventData, callback)
				}
			}

		default:
			// Not an event message - ignore (could be flight data, etc.)
			continue
		}
	}

	return nil
}

// parseEventFromRawData parses event data from raw dispatch bytes
func (sem *SystemEventManager) parseEventFromRawData(data []byte, msgType uint32) (*SystemEventData, error) {
	switch msgType {
	case SIMCONNECT_RECV_ID_EVENT:
		// Parse basic event
		event, err := ParseEvent(data)
		if err != nil {
			return nil, err
		}

		return &SystemEventData{
			EventID:   SIMCONNECT_CLIENT_EVENT_ID(event.EventID),
			EventName: "", // Will be filled by caller
			Data:      event.Data,
			EventType: "basic",
		}, nil

	case SIMCONNECT_RECV_ID_EVENT_FILENAME:
		// Parse filename event
		event, err := ParseEventFilename(data)
		if err != nil {
			return nil, err
		}
		return &SystemEventData{
			EventID:   SIMCONNECT_CLIENT_EVENT_ID(event.EventID),
			EventName: "", // Will be filled by caller
			Data:      event.Data,
			Filename:  cStringToGoString(event.SzFileName[:]),
			EventType: "filename",
		}, nil

	case SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE:
		// Parse object add/remove event
		event, err := ParseEventObjectAddRemove(data)
		if err != nil {
			return nil, err
		}

		return &SystemEventData{
			EventID:   SIMCONNECT_CLIENT_EVENT_ID(event.EventID),
			EventName: "", // Will be filled by caller
			Data:      event.Data,
			ObjectID:  event.ObjectID,
			EventType: "object",
		}, nil

	case SIMCONNECT_RECV_ID_EVENT_FRAME:
		// Parse frame event
		event, err := ParseEventFrame(data)
		if err != nil {
			return nil, err
		}

		return &SystemEventData{
			EventID:   SIMCONNECT_CLIENT_EVENT_ID(event.EventID),
			EventName: "", // Will be filled by caller
			Data:      event.Data,
			EventType: "frame",
		}, nil

	default:
		return nil, fmt.Errorf("unsupported event message type: 0x%08X", msgType)
	}
}

// processEvents processes all available system events (legacy method for compatibility)
func (sem *SystemEventManager) processEvents() error {
	for {
		// Get next event
		eventData, err := sem.client.GetSystemEvent()
		if err != nil {
			return fmt.Errorf("error getting system event: %v", err)
		}

		if eventData == nil {
			// No more events available
			break
		}

		// Find and execute callback
		sem.mutex.RLock()
		callback, exists := sem.callbacks[eventData.EventID]
		eventName, nameExists := sem.eventNames[eventData.EventID]
		sem.mutex.RUnlock()

		if exists && callback != nil {
			// Update event data with human-readable name
			if nameExists {
				eventData.EventName = eventName
			}

			// Execute callback in a separate goroutine to prevent blocking
			go func(event SystemEventData, cb SystemEventCallback) {
				defer func() {
					if r := recover(); r != nil {
						// Send panic as error to error channel
						select {
						case sem.errorChan <- fmt.Errorf("event callback panic: %v", r):
						default:
						}
					}
				}()
				cb(event)
			}(*eventData, callback)
		}
	}

	return nil
}

// SubscribeToCommonEvents is a convenience method to subscribe to commonly used events
func (sem *SystemEventManager) SubscribeToCommonEvents(callbacks map[string]SystemEventCallback) error {
	for eventName, callback := range callbacks {
		if _, err := sem.SubscribeToEvent(eventName, callback); err != nil {
			return fmt.Errorf("failed to subscribe to '%s': %v", eventName, err)
		}
	}
	return nil
}

// UnsubscribeAll unsubscribes from all events
func (sem *SystemEventManager) UnsubscribeAll() error {
	sem.mutex.Lock()
	eventIDs := make([]SIMCONNECT_CLIENT_EVENT_ID, 0, len(sem.callbacks))
	for id := range sem.callbacks {
		eventIDs = append(eventIDs, id)
	}
	sem.mutex.Unlock()

	// Unsubscribe from each event
	for _, eventID := range eventIDs {
		if err := sem.UnsubscribeFromEvent(eventID); err != nil {
			return fmt.Errorf("failed to unsubscribe from event ID %d: %v", eventID, err)
		}
	}

	return nil
}
