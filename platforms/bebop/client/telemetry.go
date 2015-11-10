package client

import (
	"bytes"
	"strconv"
	"errors"
	"encoding/json"
	"encoding/binary"
	"github.com/hybridgroup/gobot/platforms/bebop/bbtelem"
)

func (b *Bebop) Telemetry() chan bbtelem.TelemetryPacket {
	return b.telemetry
}

// Different way to handle heirarchical telemetry: attach functions to Project:Class
// doublets, then query a two-dimensional map of same using the appropriate bytes
// to get the handler function.
// Class-level switch/selection of telemetry by Id proceeds as normal.
type telemHandler struct{
  // Human readable forms of Project/Class.
  // Used as a prefix in error or unknown Id telemetry.
  ProjectName, ClassName string
  // A method on the Bebop struct that handles this telemetry event.
	// Hopefully won't have to write/desugar all my methods to functions to
	// make this work..
  HandlerFunc func(byte, *NetworkFrame)
}

// Needs Methods to call and detect errors, then issue error or unknown telemetry appropriately
// Ideally this is Lookup Map -> If Null, Issue Unknown, else Call, check errors, post errors.


func (b *Bebop) populateTelemetryHandlers() {
  b.telemetryHandlers[ARCOMMANDS_ID_PROJECT_COMMON] = map[byte]telemHandler{
    ARCOMMANDS_ID_COMMON_CLASS_COMMONSTATE: telemHandler{"Common", "CommonState", b.handleCommonStateFrame},
		ARCOMMANDS_ID_COMMON_CLASS_NETWORK: telemHandler{ "Common", "Network",
			func(a byte, f *NetworkFrame){b.sendEmptyTelemetry("networkdisconnect")}},
		ARCOMMANDS_ID_COMMON_CLASS_MAVLINKSTATE: telemHandler{"Common", "MavlinkState", b.handleMavlinkStateFrame},
		ARCOMMANDS_ID_COMMON_CLASS_CAMERASETTINGSSTATE: telemHandler{"Common", "CameraSettingsState", b.handleCameraSettingsState},
		ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANSTATE: telemHandler{"Common", "FlightPlanState", b.handleFlightPlanState},
		ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANEVENT: telemHandler{"Common", "FlightPlanEvent", b.handleFlightPlanEvent},
		ARCOMMANDS_ID_COMMON_CLASS_ARLIBSVERSIONSSTATE: telemHandler{"Common", "ARLibsVersionState", b.handleVersionStateFrames},
		ARCOMMANDS_ID_COMMON_CLASS_SETTINGSSTATE: telemHandler{"Common", "SettingsState", b.handleEventCommonSettingsState},
  }
  b.telemetryHandlers[ARCOMMANDS_ID_PROJECT_ARDRONE3] = map[byte]telemHandler{
    ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTINGSTATE: telemHandler{"ARDrone3", "PilotingState", b.handlePilotingStateFrame},
		ARCOMMANDS_ID_ARDRONE3_CLASS_CAMERASTATE: telemHandler{"ARDrone3", "CameraState", b.handleCameraStateFrame},
		ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORKSTATE: telemHandler{"ARDrone3", "NetworkState", b.handleNetworkSettingsStateFrame},
		ARCOMMANDS_ID_ARDRONE3_CLASS_PICTURESETTINGSSTATE: telemHandler{"ARDrone3", "PictureSettingsState", b.handlePictureSettingsStateFrame},
		ARCOMMANDS_ID_ARDRONE3_CLASS_GPSSETTINGSSTATE: telemHandler{"ARDrone3", "GPSSettingsState", b.handleGPSSettingsStateFrame},
  }
}

// Entry point after ACKing for data that might be worth dispatching as Telemetry.
// Hands off the work for less trivial data to other methods.
// Locate appropriate handler for the class:project and dispatch, or post a
// useful "unknown" message on failure.
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

	p_map, ok := b.telemetryHandlers[commandProject]
	if !ok {
		b.sendUnknownTelemetry("Couldn't find handlers for project: "+strconv.Itoa(int(commandProject)), frame.Data)
	}
	c_handler, ok := p_map[commandClass]
	if !ok {
		var proj string
		switch commandProject {
		case ARCOMMANDS_ID_PROJECT_COMMON:
			{proj = "Common"}
		case ARCOMMANDS_ID_PROJECT_ARDRONE3:
			{proj = "ARDrone3"}
		}
		b.sendUnknownTelemetry("Couldn't find handler for class within "+proj+": "+strconv.Itoa(int(commandProject)), frame.Data)
	}
	// TODO: Want to tie the annotation of class Handlers to their execution,
	// and perhaps reframe them to return errors instead of raising themselves,
	// for better clarity and DRYness.
  c_handler.HandlerFunc(commandId, frame)
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

// Common task: Decode a 4-byte enum, then use it to index some strings representing enum value
var enumOutOfRangeError error = errors.New("Enum value fell outside expected range.")
var enumBadSizeError error = errors.New("Wrong size binary given for a Bebop enum. Expected 4.")

func decodeEnum(raw []byte, vals []string) (string, error) {
	var evalue int
	if len(raw) != 4 {
		return "", enumBadSizeError
	}
	binary.Read(bytes.NewReader(raw), binary.LittleEndian, &evalue)
	if evalue < 0 || evalue > len(vals)-1 {
		return "", enumOutOfRangeError
	}
	return vals[evalue], nil
}

// Given a frame presumed to contain a null-ended string, return the string
// and the reamining bytes after the null
var notStringError = errors.New("Could not locate a NUL byte in presumed NUL-terminated string")

func parseNullTermedString(dataframe []byte) (string, []byte, error) {
	nul := bytes.IndexByte(dataframe, byte(0))
	if nul == -1 {
		return "", nil, notStringError
	}
	return string(dataframe[:nul]), dataframe[nul+1:], nil
}
