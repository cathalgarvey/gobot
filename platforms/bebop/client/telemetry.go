package client

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/hybridgroup/gobot/platforms/bebop/bbtelem"
	"strconv"
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

// Handles events about the camera. These seem to mostly be confirmation of user-set
// camera parameters.
func (b *Bebop) handlePictureSettingsStateFrame(commandId byte, frame *NetworkFrame) {
	switch commandId {
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_PICTUREFORMATCHANGED:
		{
			types, err := decodeEnum(frame.Data[4:8], []string{"raw", "jpeg", "snapshot"})
			if err == nil {
				payload, _ := json.Marshal(struct {
					Type string `json:"type"`
				}{Type: types})
				go b.sendTelemetry("pictureformatchanged", payload)
			} else {
				go b.sendRuntimeError("Error in pictureformatchanged handler.", err, frame.Data)
			}
		}
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_AUTOWHITEBALANCECHANGED:
		{
			types, err := decodeEnum(frame.Data[4:8], []string{"auto", "tungsten", "daylight", "cloudy", "cool_white"})
			if err == nil {
				payload, _ := json.Marshal(struct {
					Type string `json:type`
				}{Type: types})
				go b.sendTelemetry("autowhitebalancechanged", payload)
			} else {
				go b.sendRuntimeError("Error in autowhitebalancechanged handler.", err, frame.Data)
			}
		}
		// Handle Exposition / Saturation identically except for telemetry dispatch name.
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_EXPOSITIONCHANGED,
		ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_SATURATIONCHANGED:
		{
			var telemdata struct {
				Value float32 `json:"value"`
				Min   float32 `json:"min"`
				Max   float32 `json:"max"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			payload, err := json.Marshal(telemdata)
			if err != nil {
				go b.sendRuntimeError("Error in Saturation/Exposition telemetry handler", err, frame.Data)
			}
			switch commandId {
			case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_EXPOSITIONCHANGED:
				{
					go b.sendTelemetry("expositionchanged", payload)
				}
			case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_SATURATIONCHANGED:
				{
					go b.sendTelemetry("saturationchanged", payload)
				}
			}
		}
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_TIMELAPSECHANGED:
		{
			var telemdata struct {
				Enabled     bool    `json:"enabled"`
				Interval    float32 `json:"interval"`
				MinInterval float32 `json:"minInterval"`
				MaxInterval float32 `json:"maxInterval"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:17]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("timelapsechanged", payload)
		}
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_VIDEOAUTORECORDCHANGED:
		{
			var telemdata struct {
				Enabled         bool  `json:"enabled"`
				Mass_storage_id uint8 `json:"mass_storage_id"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:6]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("videoautorecordchanged", payload)
		}
	default:
		{
			go b.sendUnknownTelemetry("Unknown picture settings commandId: "+strconv.Itoa(int(commandId)), frame.Data)
		}
	}
}

// Handle telemetry from device pertaining to Wifi band/channel settings
func (b *Bebop) handleNetworkSettingsStateFrame(commandId byte, frame *NetworkFrame) {
	switch commandId {
	case ARCOMMANDS_ARDRONE3_NETWORKSETTINGSSTATECHANGED_STATE_WIFISELECTIONCHANGED:
		// Appears to be simply feedback for when the client issues a corresponding
		// instruction; returns settings to confirm?
		{
			wftypestr, err := decodeEnum(frame.Data[4:8], []string{"auto_all", "auto_2_4ghz", "auto_5ghz", "all"})
			if err != nil {
				go b.sendRuntimeError("Error in WIFISELECTIONCHANGED telemetry handler", err, frame.Data)
				return
			}
			wfbandstr, err := decodeEnum(frame.Data[8:12], []string{"2_4ghz", "5ghz", "all"})
			if err != nil {
				go b.sendRuntimeError("Error in WIFISELECTIONCHANGED telemetry handler", err, frame.Data)
				return
			}
			var channel uint8
			binary.Read(bytes.NewReader(frame.Data[12:13]), binary.LittleEndian, &channel)
			payload, _ := json.Marshal(struct {
				Type    string `json:"type"`
				Band    string `json:"band"`
				Channel int    `json:"channel"`
			}{Type: wftypestr, Band: wfbandstr, Channel: int(channel)})
			go b.sendTelemetry("networksettingsstate", payload)
		}
	default:
		{
			go b.sendUnknownTelemetry("Unknown Network commandId: "+strconv.Itoa(int(commandId)), frame.Data)
		}
	}
}

// Handles the important events that related to device state in the air: GPS position,
// attitude, speed, etcetera.
func (b *Bebop) handlePilotingStateFrame(commandId byte, frame *NetworkFrame) {
	switch commandId {
	// Flat Trim changed (?)
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_FLATTRIMCHANGED:
		{
			// No args. Very often.
			go b.sendEmptyTelemetry("flattrim")
		}
	// Flying state changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_FLYINGSTATECHANGED:
		{
			var flyingstate int
			binary.Read(bytes.NewReader(frame.Data[4:8]), binary.LittleEndian, &flyingstate)
			// These are kind of a big deal so send them as separate events, unlike other enums
			switch byte(flyingstate) {
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_LANDED:
				{
					go b.sendEmptyTelemetry("landed")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_TAKINGOFF:
				{
					go b.sendEmptyTelemetry("takingoff")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_HOVERING:
				{
					go b.sendEmptyTelemetry("hovering")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_FLYING:
				{
					go b.sendEmptyTelemetry("flying")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_LANDING:
				{
					go b.sendEmptyTelemetry("landing")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_EMERGENCY:
				{
					go b.sendEmptyTelemetry("emergency")
				}
			}
		}
	// Alert State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALERTSTATECHANGED:
		{
			statestr, err := decodeEnum(frame.Data[4:8], []string{"none", "cut_out", "critical_battery", "low_battery", "too_much_angle"})
			if err != nil {
				go b.sendRuntimeError("Error in ALERTSTATECHANGED telemetry handler", err, frame.Data)
				return
			}
			payload, _ := json.Marshal(struct {
				State string `json:"state"`
			}{State: statestr})
			go b.sendTelemetry("alertstate", payload)
		}
	// Navigate Home State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_NAVIGATEHOMESTATECHANGED:
		{
			statestr, err := decodeEnum(frame.Data[4:8], []string{"available", "inProgress", "unavailable", "pending"})
			if err != nil {
				go b.sendRuntimeError("Error in NAVIGATEHOMESTATECHANGED telemetry handler", err, frame.Data)
				return
			}
			reasonstr, err := decodeEnum(frame.Data[8:12], []string{"userRequest", "connectionLost", "lowBattery", "finished", "stopped", "disabled", "enabled"})
			if err != nil {
				go b.sendRuntimeError("Error in NAVIGATEHOMESTATECHANGED telemetry handler", err, frame.Data)
				return
			}
			payload, _ := json.Marshal(struct {
				State  string `json:"state"`
				Reason string `json:"reason"`
			}{
				State: statestr, Reason: reasonstr,
			})
			go b.sendTelemetry("navigatehomestate", payload)
		}
	// Position (GPS)
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_POSITIONCHANGED:
		{
			var telemdata struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
				Alt float64 `json:"alt"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:28]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("gps", payload)
		}
	// Speed Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_SPEEDCHANGED:
		{
			var telemdata struct {
				SpeedX float64 `json:"speedX"`
				SpeedY float64 `json:"speedY"`
				SpeedZ float64 `json:"speedZ"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:28]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("speed", payload)
		}
	// Attitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ATTITUDECHANGED:
		{
			var telemdata struct {
				Roll  float32 `json:"roll"`
				Pitch float32 `json:"pitch"`
				Yaw   float32 `json:"yaw"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("attitude", payload)
		}
	// Auto Takeoff Mode Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_AUTOTAKEOFFMODECHANGED:
		{
			var telemdata struct {
				State bool `json:"state"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("autotakeoffmode", payload)
		}
	// Altitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALTITUDECHANGED:
		{
			var telemdata struct {
				Altitude float64 `json:"altitude"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:12]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("altitude", payload)
		}
	// End of PilotingState cases
	default:
		{
			go b.sendUnknownTelemetry("Unknown Piloting State", frame.Data)
		}
	}
}

func (b *Bebop) handleCommonStateFrame(commandId byte, frame *NetworkFrame) {
	switch commandId {
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_ALLSTATESCHANGED:
		{
			// Is this useful telemetry?
			return
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_BATTERYSTATECHANGED:
		{
			// This uint8 is a percentage acc. to docs, should be 0-100?
			var telemdata struct {
				Battery uint8 `json:"battery"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("battery", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGESTATELISTCHANGED:
		{
			var (
				mass_storage_id uint8
			)
			binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &mass_storage_id)
			mass_storage_name := string(frame.Data[5:]) // ? Encoding? Length? Huh?
			payload, _ := json.Marshal(struct {
				Mass_storage_id uint8  `json:"Mass_storage_id"`
				Name            string `json:"name"`
			}{Mass_storage_id: mass_storage_id, Name: mass_storage_name})
			go b.sendTelemetry("massstorage", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGEINFOSTATELISTCHANGED:
		// Information on a particular volume? Volunteered, or in response to a query?
		{
			var telemdata struct {
				Mass_storage_id uint8  `json:"mass_storage_id"`
				Size            uint32 `json:"size"`
				Used_size       uint32 `json:"used_size"`
				Plugged         bool   `json:"plugged"`
				Full            bool   `json:"full"`
				Internal        bool   `json:"internal"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:72]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("massstorageinfo", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTDATECHANGED:
		// Date in ISO-8601
		{
			dates := string(frame.Data[4:]) // Parse to real time object? ISO-8601
			payload, _ := json.Marshal(struct {
				Date string `json:"date"`
			}{Date: dates})
			go b.sendTelemetry("currentdate", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTTIMECHANGED:
		// Time in ISO-8601
		{
			times := string(frame.Data[4:]) // Parse to real time object? ISO-8601
			payload, _ := json.Marshal(struct {
				Time string `json:"time"`
			}{Time: times})
			go b.sendTelemetry("currenttime", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGEINFOREMAININGLISTCHANGED:
		// Remaining space on volume, with estimate of photo space/recording time?
		{
			var telemdata struct {
				Free_space      uint32 `json:"free_space"`
				Rec_time        uint16 `json:"rec_time"`
				Photo_remaining uint32 `json:"photo_remaining"`
			}
			binary.Read(bytes.NewReader(frame.Data[4:80]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("massstorageinforemaining", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_WIFISIGNALCHANGED:
		{
			var telemdata struct {
				Rssi int16 `json:"rssi"`
			} // in dbm
			binary.Read(bytes.NewReader(frame.Data[4:20]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("wifisignal", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_SENSORSSTATESLISTCHANGED:
		{
			var sensorState bool
			sensorName, err := decodeEnum(frame.Data[4:8], []string{"IMU", "barometer", "ultrasound", "GPS", "magnetometer", "vertical_camera"})
			if err != nil {
				go b.sendRuntimeError("Error processing sensor state telemetry", err, frame.Data)
				return
			}
			pld := struct {
				SensorName  string `json:"sensorName"`
				SensorState bool   `json:"sensorState"`
			}{SensorName: sensorName, SensorState: sensorState}
			payload, _ := json.Marshal(pld)
			go b.sendTelemetry("sensorstates", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_PRODUCTMODEL:
		// This appears to be irrelevant to the Bebop but it's in "common"!
		{
			modelstr, err := decodeEnum(frame.Data[4:8], []string{"RS_TRAVIS", "RS_MARS", "RS_SWAT", "RS_MCLANE", "RS_BLAZE", "RS_ORAK", "RS_NEWZ", "JS_DIESEL", "JS_BUZZ", "JS_MAX", "JS_JETT", "JS_TUKTUK"})
			if err != nil {
				go b.sendRuntimeError("Error processing drone model telemetry", err, frame.Data)
				return
			}
			payload, _ := json.Marshal(struct {
				Model string `json:"model"`
			}{Model: modelstr})
			go b.sendTelemetry("dronemodel", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_COUNTRYLISTKNOWN:
		{
			ccodes := string(frame.Data[4:])
			payload, _ := json.Marshal(struct {
				CountryCodes string `json:"countryCodes"`
			}{ccodes})
			go b.sendTelemetry("countrycodes", payload)
		}
	default:
		{
			go b.sendUnknownTelemetry("Unknown/Unhandled COMMONSTATE commandId: "+strconv.Itoa(int(commandId)), frame.Data)
		}
	}
}

func (b *Bebop) handleVersionStateFrames(commandId byte, frame *NetworkFrame) {
	switch commandId {
	case 0: // ControllerLibARCommandsVersion
		{
			version, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				b.sendRuntimeError("Error parsing controller libARCCommands version frame", err, frame.Data)
				return
			}
			payload, _ := json.Marshal(struct{ Version string }{Version: version})
			go b.sendTelemetry("controllerlibversion", payload)
		}
	case 1: // SkyControllerLibARCommandsVersion
		{
			version, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				b.sendRuntimeError("Error parsing skycontroller libARCCommands version frame", err, frame.Data)
				return
			}
			payload, _ := json.Marshal(struct{ Version string }{Version: version})
			go b.sendTelemetry("skycontrollerlibversion", payload)
		}
	case 2: // DeviceLibARCommandsVersion
		{
			version, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				b.sendRuntimeError("Error parsing device libARCCommands version frame", err, frame.Data)
				return
			}
			payload, _ := json.Marshal(struct{ Version string }{Version: version})
			go b.sendTelemetry("devicelibversion", payload)
		}
	}
}

// Handle common Mavlink/Flightplan state frame
func (b *Bebop) handleMavlinkStateFrame(commandId byte, frame *NetworkFrame) {
	switch commandId {
	case 0: // MavlinkFilePlayingStateChanged,  Playing state of a mavlink flight plan
		{
			state, err := decodeEnum(frame.Data[4:8], []string{"playing", "stopped", "paused"})
			if err != nil {
				b.sendRuntimeError("Error decoding 'state' from MavlinkFilePlayingStateChanged frame", err, frame.Data)
				return
			}
			filepath, rest, err := parseNullTermedString(frame.Data[8:])
			if err != nil {
				b.sendRuntimeError("Error decoding 'filepath' from MavlinkFilePlayingStateChanged frame", err, frame.Data)
				return
			}
			types, err := decodeEnum(rest, []string{"flightPlan", "mapMyHouse"})
			payload, _ := json.Marshal(struct {
				State    string `json:"state"`
				Filepath string `json:"filepath"`
				Type     string `json:"type"`
			}{state, filepath, types})
			go b.sendTelemetry("mavlinkfileplaying", payload)
		}
	}
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

// Entry point after ACKing for data that might be worth dispatching as Telemetry.
// Hands off the work for less trivial data to other methods.
func (b *Bebop) handleIncomingDataFrame(frame *NetworkFrame) {
	var (
		commandProject byte
		commandClass   byte
		commandId      byte
		commandId16    uint16
	)
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

				}
			case ARCOMMANDS_ID_COMMON_CLASS_MAVLINKSTATE:
				{
					// Refers to a CSV.
					b.handleMavlinkStateFrame(commandId, frame)
				}
			case ARCOMMANDS_ID_COMMON_CLASS_CAMERASETTINGSSTATE:
				{
					// Appears in static log
				}
			case ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANSTATE:
				{
					// Dumps regularly
				}
			case ARCOMMANDS_ID_COMMON_CLASS_FLIGHTPLANEVENT:
				{
					//
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
					payload, _ := json.Marshal(telemdata)
					go b.sendTelemetry("camerastate", payload)
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
			go b.sendUnknownTelemetry("Unknown Project: "+strconv.Itoa(int(commandProject)), frame.Data)
		}
	}
}
