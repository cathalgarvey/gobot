package client

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hybridgroup/gobot/platforms/bebop/bbtelem"
)

// Telemetry returns a channel on which to receive telemetry packets.
func (b *Bebop) Telemetry() chan bbtelem.TelemetryPacket {
	return b.telemetry
}

// Different way to handle heirarchical telemetry: attach functions to Project:Class
// doublets, then query a two-dimensional map of same using the appropriate bytes
// to get the handler function.
// Class-level switch/selection of telemetry by Id proceeds as normal.
type telemHandler struct {
	// Human readable forms of Project/Class.
	// Used as a prefix in error or unknown Id telemetry.
	ProjectName, ClassName string
	// A method on the Bebop struct that handles this telemetry event.
	// Returns true if a handler was found, and error if the handler broke.
	// This allows the dispatcher to use the above naming information to send
	// "unknown" or "error" events.
	HandlerFunc func(byte, *NetworkFrame) (bool, string, error)
}

// Needs Methods to call and detect errors, then issue error or unknown telemetry appropriately
// Ideally this is Lookup Map -> If Null, Issue Unknown, else Call, check errors, post errors.

func (b *Bebop) populateTelemetryHandlers() {
	b.telemetryHandlers[ARCOMMANDS_ID_PROJECT_COMMON] = map[byte]telemHandler{
		ARCOMMANDS_ID_COMMON_CLASS_COMMONSTATE:         telemHandler{"Common", "CommonState", b.handleCommonStateFrame},
		ARCOMMANDS_ID_COMMON_CLASS_NETWORK:             telemHandler{"Common", "Network", b.handleNetworkFrame},
		ARCOMMANDS_ID_COMMON_CLASS_MAVLINKSTATE:        telemHandler{"Common", "MavlinkState", b.handleMavlinkStateFrame},
		ARCOMMANDS_ID_COMMON_CLASS_CAMERASETTINGSSTATE: telemHandler{"Common", "CameraSettingsState", b.handleCameraSettingsState},
		ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANSTATE:     telemHandler{"Common", "FlightPlanState", b.handleFlightPlanState},
		ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANEVENT:     telemHandler{"Common", "FlightPlanEvent", b.handleFlightPlanEvent},
		ARCOMMANDS_ID_COMMON_CLASS_ARLIBSVERSIONSSTATE: telemHandler{"Common", "ARLibsVersionState", b.handleVersionStateFrames},
		ARCOMMANDS_ID_COMMON_CLASS_SETTINGSSTATE:       telemHandler{"Common", "SettingsState", b.handleEventCommonSettingsState},
	}
	b.telemetryHandlers[ARCOMMANDS_ID_PROJECT_ARDRONE3] = map[byte]telemHandler{
		// 0  = ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTING -> Command!
		// 1: Why is this returning telemetry? These are camera controlling commands? -> Command!
		ARCOMMANDS_ID_ARDRONE3_CLASS_CAMERA: telemHandler{"ARDrone3", "Camera", b.handleCameraFrame},
		// 2  = ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTINGSETTINGS -> Command! Sets maxima/minima, V. important to implement.
		// 3  = ARCOMMANDS_ID_ARDRONE3_CLASS_MEDIARECORDEVENT
		// 4:
		ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTINGSTATE: telemHandler{"ARDrone3", "PilotingState", b.handlePilotingStateFrame},
		// 5  = ARCOMMANDS_ID_ARDRONE3_CLASS_ANIMATIONS
		// 6  = ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTINGSETTINGSSTATE
		// TODO: Implement! Reports results of <2>?
		ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTINGSETTINGSSTATE: telemHandler{"ARDrone3", "PilotingSettingsState", b.handlePilotingSettingsState},
		// 7  = ARCOMMANDS_ID_ARDRONE3_CLASS_MEDIARECORD
		// 8  = ARCOMMANDS_ID_ARDRONE3_CLASS_MEDIARECORDSTATE
		// 9  = ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORKSETTINGS
		// 10 = ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORKSETTINGSSTATE
		ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORKSETTINGSSTATE: telemHandler{"ARDrone3", "NetworkSettingsState", b.handleNetworkSettingsStateFrame},
		// 11 = ARCOMMANDS_ID_ARDRONE3_CLASS_SPEEDSETTINGS
		// 12: = ARCOMMANDS_ID_ARDRONE3_CLASS_SPEEDSETTINGSSTATE
		ARCOMMANDS_ID_ARDRONE3_CLASS_SPEEDSETTINGSSTATE: telemHandler{"ARDrone3", "SpeedSettingsState", b.handleSpeedSettingsState},
		// 13 = ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORK  -> Command! Responses are via 14.
		// 14:
		ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORKSTATE: telemHandler{"ARDrone3", "NetworkState", b.handleNetworkStateFrame},
		// 15 = ARCOMMANDS_ID_ARDRONE3_CLASS_SETTINGS              byte = 15
		// 16 = ARCOMMANDS_ID_ARDRONE3_CLASS_SETTINGSSTATE         byte = 16
		// 17 = ARCOMMANDS_ID_ARDRONE3_CLASS_DIRECTORMODE          byte = 17
		// 18 = ARCOMMANDS_ID_ARDRONE3_CLASS_DIRECTORMODESTATE     byte = 18
		// 19 = ARCOMMANDS_ID_ARDRONE3_CLASS_PICTURESETTINGS       byte = 19
		// 20:
		ARCOMMANDS_ID_ARDRONE3_CLASS_PICTURESETTINGSSTATE: telemHandler{"ARDrone3", "PictureSettingsState", b.handlePictureSettingsStateFrame},
		// 21 = ARCOMMANDS_ID_ARDRONE3_CLASS_MEDIASTREAMING        byte = 21
		// 22 = ARCOMMANDS_ID_ARDRONE3_CLASS_MEDIASTREAMINGSTATE   byte = 22
		// 23 = ARCOMMANDS_ID_ARDRONE3_CLASS_GPSSETTINGS           byte = 23
		// 24:
		ARCOMMANDS_ID_ARDRONE3_CLASS_GPSSETTINGSSTATE: telemHandler{"ARDrone3", "GPSSettingsState", b.handleGPSSettingsStateFrame},
		// 25:
		ARCOMMANDS_ID_ARDRONE3_CLASS_CAMERASTATE: telemHandler{"ARDrone3", "CameraState", b.handleCameraStateFrame},
		// 29 = ARCOMMANDS_ID_ARDRONE3_CLASS_ANTIFLICKERING        byte = 29
		// 30 = ARCOMMANDS_ID_ARDRONE3_CLASS_ANTIFLICKERINGSTATE   byte = 30
		// 31 = ?
		// 32 = ?
		// 33 = ?
		// 34 = PilotingEvent, not yet implemented, represents response to Piloting relative-movement instruction
	}
}

// Entry point after ACKing for data that might be worth dispatching as Telemetry.
// Hands off the work for less trivial data to other methods.
// Locate appropriate handler for the class:project and dispatch, or post a
// useful "unknown" message on failure.
func (b *Bebop) handleIncomingDataFrame(frame *NetworkFrame) {
	var (
		commandProject byte // Seems to increment continuously on some frames?
		commandClass   byte
		commandID16    uint16
		commandID      byte
	)
	// For single-byte values is this overkill?
	binary.Read(bytes.NewReader(frame.Data[0:1]), binary.LittleEndian, &commandProject)
	binary.Read(bytes.NewReader(frame.Data[1:2]), binary.LittleEndian, &commandClass)
	binary.Read(bytes.NewReader(frame.Data[2:4]), binary.LittleEndian, &commandID16)
	commandID = byte(commandID16)

	pMap, ok := b.telemetryHandlers[commandProject]
	if !ok {
		b.sendUnknownTelemetry("Couldn't find handlers for project: "+strconv.Itoa(int(commandProject)), frame.Data)
	}
	cHandler, ok := pMap[commandClass]
	if !ok {
		var proj string
		switch commandProject {
		case ARCOMMANDS_ID_PROJECT_COMMON:
			{
				proj = "Common"
			}
		case ARCOMMANDS_ID_PROJECT_ARDRONE3:
			{
				proj = "ARDrone3"
			}
		}
		b.sendUnknownTelemetry("Couldn't find handler for class within "+proj+": "+strconv.Itoa(int(commandProject)), frame.Data)
		return
	}
	// Handlers return (true, nil) if everything went well, or (false, nil) if
	// handler for the commandID wasn't found, or (true, error) if the handler
	// broke somehow.
	// Context value is only used when the commandID was located; it is a human-readable
	// reference to the command.
	go func(cHandler *telemHandler, commandID byte, frame *NetworkFrame) {
		path := cHandler.ProjectName + ":" + cHandler.ClassName
		cmdidstr := strconv.Itoa(int(commandID))
		found, context, err := cHandler.HandlerFunc(commandID, frame)
		if err != nil {
			b.sendRuntimeError("Error in handler for "+path+", commandID "+cmdidstr+", context '"+context+"'", err, frame.Data)
		}
		if !found {
			b.sendUnknownTelemetry("Unknown commandID in "+path+": "+cmdidstr, frame.Data)
		}
	}(&cHandler, commandID, frame)
}

// ErrTelemSend occurs in dispatchTelemetry, an internal method, and may be propagated. It is sent
// when the internal channel for telemetry is full.
var ErrTelemSend = errors.New("Failed to send telemetry; channel full.")

// Attempts to send data with a given title across the telemetry channel. If the
// chan is full then the default simply drops the data.
func (b *Bebop) dispatchTelemetry(telem *bbtelem.TelemetryPacket) error {
	select {
	case <-b.endTelemetry:
		{
			return nil
		}
	case b.telemetry <- *telem:
		{
			return nil
		}
	// If buffer above is full (10 unread), abandon send.
	default:
		{
			return ErrTelemSend
		}
	}
}

// Make a telemetry object and ship to dispatchTelemetry
func (b *Bebop) sendTelemetry(title string, payload []byte) error {
	return b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title:   title,
		Payload: payload,
	})
}

// Shortcut method for sending a title with an empty JSON object.
func (b *Bebop) sendEmptyTelemetry(title string) error {
	return b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title: title,
	})
}

// Shortcut method for sending unknown data embedded in a JSON object as {"data": "<base64>"}
func (b *Bebop) sendUnknownTelemetry(comment string, data []byte) error {
	return b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title:   bbtelem.Unknown, // "unknown",
		Comment: comment,
		Payload: data,
	})
}

// Shortcut method for issuing errors through Telemetry
func (b *Bebop) sendRuntimeError(comment string, err error, data []byte) error {
	internalErr := b.dispatchTelemetry(&bbtelem.TelemetryPacket{
		Title:   bbtelem.Error, // "error",
		Error:   err,
		Comment: comment,
		Payload: data,
	})
	if internalErr != nil {
		fmt.Println("RUNTIME ERROR: ", internalErr.Error())
	}
	return internalErr
}

// Handles the very common job of encoding to JSON, while handling errors. Errors
// are currently silently ignored, using this will help handle them well without
// imposing code overhead or duplication.
func (b *Bebop) sendJSONTelemetry(frame *NetworkFrame, eventTitle string, obj interface{}) error {
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return b.sendTelemetry(eventTitle, payload)
}

// StopTelemetry stops further telemetry from being sent
func (b *Bebop) StopTelemetry() error {
	close(b.endTelemetry)
	return nil
}

// ErrEnumOutOfRange can be returned by decodeEnum, an internal function, and may be propagated
// elsewhere.
var ErrEnumOutOfRange = errors.New("Enum value fell outside expected range.")

// ErrEnumBadSize can be returned by decodeEnum, an internal function, and may be propagated
// elsewhere.
var ErrEnumBadSize = errors.New("Wrong size binary given for a Bebop enum. Expected 4.")

// Common task: Decode a 4-byte enum, then use it to index some strings representing enum value
func decodeEnum(raw []byte, vals []string) (string, error) {
	var (
		evalue  uint32
		evaluei int
	)
	if len(raw) != 4 {
		return "", ErrEnumBadSize
	}
	err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, &evalue)
	if err != nil {
		return "", err
	}
	evaluei = int(evalue)
	if evaluei < 0 || evaluei > len(vals)-1 {
		return "", ErrEnumOutOfRange
	}
	return vals[evaluei], nil
}

// ErrNotString is returned if a string expected to end in NUL, didn't (returned from internal function parseNullTermedString).
var ErrNotString = errors.New("Could not locate a NUL byte in presumed NUL-terminated string")

//Given a frame presumed to contain a null-ended string, return the string
// and the reamining bytes after the null
func parseNullTermedString(dataframe []byte) (string, []byte, error) {
	nul := bytes.IndexByte(dataframe, byte(0))
	if nul == -1 {
		return "", nil, ErrNotString
	}
	return string(dataframe[:nul]), dataframe[nul+1:], nil
}
