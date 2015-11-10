package client

import (
	"bytes"
	"strconv"
	"encoding/json"
	"encoding/binary"
	"github.com/hybridgroup/gobot/platforms/bebop/bbtelem"
)

func (b *Bebop) Telemetry() chan bbtelem.TelemetryPacket {
	return b.telemetry
}

// Attempts to send data with a given title across the telemetry channel. If the
// chan is full then the default simply drops the data.
func (b *Bebop) dispatchTelemetry(telem *bbtelem.TelemetryPacket) {
	select {
	case <-b.endTelemetry:
		{
			return
		}
	case b.telemetry <- *telem:
		{
			return
		}
	// If buffer above is full (10 unread), abandon send.
	default:
		{
			return
		}
	}
}

// Make a telemetry object and ship to dispatchTelemetry
func (b *Bebop) sendTelemetry(title string, payload []byte) {
	b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title:   title,
		Payload: payload,
	})
}

// Shortcut method for sending a title with an empty JSON object.
func (b *Bebop) sendEmptyTelemetry(title string) {
	b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title: title,
	})
}

// Shortcut method for sending unknown data embedded in a JSON object as {"data": "<base64>"}
func (b *Bebop) sendUnknownTelemetry(comment string, data []byte) {
	b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title:   "unknown",
		Comment: comment,
		Payload: data,
	})
}

// Shortcut method for issuing errors through Telemetry
func (b *Bebop) sendRuntimeError(comment string, err error, data []byte) {
	b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title:   "error",
		Error:   err,
		Comment: comment,
		Payload: data,
	})
}

// Handles the very common job of encoding to JSON, while handling errors. Errors
// are currently silently ignored, using this will help handle them well without
// imposing code overhead or duplication.
func (b *Bebop) sendJSONTelemetry(frame *NetworkFrame, eventTitle string, obj interface{}) {
  payload, err := json.Marshal(obj)
  if err != nil {
    b.sendRuntimeError("Error encoding JSON for ''"+eventTitle+"' event", err, frame.Data)
    return
  }
  go b.sendTelemetry(eventTitle, payload)
}


// Entry point after ACKing for data that might be worth dispatching as Telemetry.
// Hands off the work for less trivial data to other methods.
func (b *Bebop) handleIncomingDataFrame(frame *NetworkFrame) {
	var (
		commandProject byte  // Seems to increment continuously on some frames?
		commandClass   byte
		commandId16    uint16
		commandId      byte
	)
	// For single-byte values is this overkill?
	binary.Read(bytes.NewReader(frame.Data[0:1]), binary.LittleEndian, &commandProject)
	binary.Read(bytes.NewReader(frame.Data[1:2]), binary.LittleEndian, &commandClass)
	binary.Read(bytes.NewReader(frame.Data[2:4]), binary.LittleEndian, &commandId)
	commandId = byte(commandId16)

	switch commandProject {
	case ARCOMMANDS_ID_PROJECT_COMMON:
		{
			switch commandClass {
			case ARCOMMANDS_ID_COMMON_CLASS_NETWORK:
				{
					// Urgent: dumps regularly.
					// Only has one command, "Disconnect"
					go b.sendEmptyTelemetry("networkdisconnect")
				}
			case ARCOMMANDS_ID_COMMON_CLASS_MAVLINKSTATE:
				{
					// Refers to a CSV.
					b.handleMavlinkStateFrame(commandId, frame)
				}
			case ARCOMMANDS_ID_COMMON_CLASS_CAMERASETTINGSSTATE:
				{
					// Appears in static log
					// Only one command, "camerasettingsstate"
					var telemdata struct{
						Fov float32 `json:"fov"`
						PanMax float32 `json:"panMax"`
						PanMin float32 `json:"panMin"`
						TiltMax float32 `json:"tiltMax"`
						TileMin float32 `json:"tileMin"`
					}
					binary.Read(bytes.NewReader(frame.Data[4:4+(32 * 5)]), binary.LittleEndian, &telemdata)
					b.sendJSONTelemetry(frame, "camerasettingsstate", telemdata)
				}
			case ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANSTATE:
				{
					// Dumps regularly
					// One command "AvailabilityStateChanged"
					var telemdata struct{
						AvailabilityState bool `json:"availabilityState"`
					}
					binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
					b.sendJSONTelemetry(frame, "availabilitystatechanged", telemdata)
				}
			case ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANEVENT:
				{
					switch commandId {
					case 0: // StartingErrorEvent - Event of flight plan start error
						{
							b.sendEmptyTelemetry("startingerrorevent")
						}
					case 1: // SpeedBridleEvent - Bridle speed of the drone
						{
						  b.sendEmptyTelemetry("speedbridleevent")
						}
					}
				}
			case ARCOMMANDS_ID_COMMON_CLASS_ARLIBSVERSIONSSTATE:
				{
					// Dumps early, seems to just be a version number.
				  b.handleVersionStateFrames(commandId, frame)
				}
			case ARCOMMANDS_ID_COMMON_CLASS_COMMONSTATE:
				{
					b.handleCommonStateFrame(commandId, frame)
				}
			default:
				{
					go b.sendUnknownTelemetry("Unknown/Unhandled common project commandClass: "+strconv.Itoa(int(commandClass)), frame.Data)
				}
			}
		}
	case ARCOMMANDS_ID_PROJECT_ARDRONE3:
		{
			switch commandClass {
			case ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTINGSTATE:
				// This includes things like speed, altitude, GPS coords, and current
				// gross behaviour ("flying"/"landing").
				{
					b.handlePilotingStateFrame(commandId, frame)
				}
			case ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORKSTATE:
				// Not as interesting as it sounds, this handleds onboard settings
				// around frequency and channel use.
				{
					b.handleNetworkSettingsStateFrame(commandId, frame)
				}
			case ARCOMMANDS_ID_ARDRONE3_CLASS_PICTURESETTINGSSTATE:
				// Appears to be simply feedback for when the client issues a corresponding
				// instruction; returns settings to confirm?
				{
					b.handlePictureSettingsStateFrame(commandId, frame)
				}
			case ARCOMMANDS_ID_ARDRONE3_CLASS_CAMERASTATE:
				{
					// Only one commandId, 0. Don't bother checking?
					var telemdata struct{ Tilt, Pan int8 }
					binary.Read(bytes.NewReader(frame.Data[4:6]), binary.LittleEndian, &telemdata)
					b.sendJSONTelemetry(frame, "camerastate", telemdata)
				}
			case ARCOMMANDS_ID_ARDRONE3_CLASS_GPSSETTINGSSTATE:
				{
					// command Class 24, gets dumped early on Gobot init sequence.
					b.handleGPSSettingsStateFrame(commandId, frame)
				}
			default:
				{
					go b.sendUnknownTelemetry("Unknown/Unhandled ARDRONE3 command Class: "+strconv.Itoa(int(commandClass)), frame.Data)
				}
			}
		}
	default:
		{
			// This shouldn't happen, as there are only two expected projects?
			// Post an unknown telemetry event, may help to discover stuff for future usage.
			go b.dispatchTelemetry(&bbtelem.TelemetryPacket{
				Title:   "unknownProject",
				Comment: "Unknown commandProject: "+strconv.Itoa(int(commandProject)),
				Payload: frame.Data,
			})
		}
	}
}
