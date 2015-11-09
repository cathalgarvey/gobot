package client

import (
	"bytes"
	"errors"
	"encoding/json"
	"encoding/binary"
)

// Attempts to send data with a given title across the telemetry channel. If the
// chan is full then the default simply drops the data.
func (b *Bebop) sendTelemetry(title string, data []byte) {
	select {
	case <-b.endTelemetry:
		{
			return
		}
	case b.telemetry <- struct {
		Title string
		Data  []byte
	}{Title: title, Data: data}:
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

// Shortcut method for sending a title with an empty JSON object.
func (b *Bebop) sendEmptyTelemetry(title string) {
	sendTelementry(title, []byte("{}"))
}

// Shortcut method for sending unknown data embedded in a JSON object as {"data": "<base64>"}
func (b *Bebop) sendUnknownTelemetry(comment string, data []byte) {
	payload, _ := json.Marshal(struct{data []byte; comment string}{data: data, comment: comment})
	sendTelemetry("unknown", payload)
}

// Shortcut method for issuing errors through Telemetry
func (b *Bebop) sendRuntimeError(comment string, err error, data []byte) {
	payload, _ := json.Marshal(struct{data []byte; comment, error string}{data: data, comment: comment, error: err.Error()})
	sendTelemetry("error", payload)
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
		case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_PICTUREFORMATCHANGED: {
		  types, err := decodeEnum(frame.data[4:8], []string{"raw", "jpeg", "snapshot"})
			if err == nil {
				payload, _ := json.Marshal(struct{type string}{type: types})
				go sendTelemetry("pictureformatchanged", payload)
			} else {
				go sendRuntimeError("Error in pictureformatchanged handler.", err, frame.data)
			}
		}
		case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_AUTOWHITEBALANCECHANGED: {
			types, err := decodeEnum(frame.data[4:8], []string{"auto", "tungsten", "daylight", "cloudy", "cool_white"})
			if err == nil {
				payload, _ := json.Marshal(struct{type string}{type: types})
				go sendTelemetry("autowhitebalancechanged", payload)
			} else {
				go sendRuntimeError("Error in autowhitebalancechanged handler.", err, frame.data)
			}
		}
    // Handle Exposition / Saturation identically except for telemetry dispatch name.
		case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_EXPOSITIONCHANGED,
				 ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_SATURATIONCHANGED: {
			var telemdata struct{value, min, max float32}
			binary.Read(bytes.NewReader(frame.data[ 4: 16]), binary.LittleEndian, &telemdata)
			payload, err := json.Marshal(telemdata)
			if err != nil {
				go sendRuntimeError("Error in Saturation/Exposition telemetry handler", err, frame.data)
			}
			switch commandId {
			case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_EXPOSITIONCHANGED: {
				go sendTelemetry("expositionchanged", payload)
			}
			case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_SATURATIONCHANGED: {
			  go sendTelemetry("saturationchanged", payload)
			}
		}
		case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_TIMELAPSECHANGED: {
			var (
				enabled byte
				interval, minInterval, maxInterval float32
			)
			var telemdata struct{enabled bool; interval, minInterval, maxInterval float32}
			binary.Read(bytes.NewReader(frame.data[ 4: 17]), binary.LittleEndian, &telemdata)
			payload, err := json.Marshal(telemdata)
			if err != nil {
				go sendRuntimeError("Error in timelapsechanged telemetry", err, frame.data)
				return
			}
			go sendTelemetry("timelapsechanged", payload)
		}
		case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_VIDEOAUTORECORDCHANGED: {
			var telemdata struct{enabled bool; mass_storage_id uint8}
			binary.Read(bytes.NewReader(frame.data[ 4: 6]), binary.LittleEndian, &telemdata)
			payload, err := json.Marshal(telemdata)
			if err != nil {
				go sendRuntimeError("Error in videoautorecordchanged telemetry", err, frame.data)
				return
			}
			go sendTelemetry("videoautorecordchanged", payload)
		}
		default: {
			go sendUnknownTelemetry("Unknown picture settings commandId: "+strconv.Itoa(int(commandId)), frame.data)
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
			wftypestr, err := decodeEnum(frame.data[4:8], []string{"auto_all", "auto_2_4ghz", "auto_5ghz", "all"})
			if err != nil {
				go sendRuntimeError("Error in WIFISELECTIONCHANGED telemetry handler", err, frame.data)
				return
			}
			wfbandstr, err := decodeEnum(frame.data[8:12], []string{"2_4ghz", "5ghz", "all"})
			if err != nil {
				go sendRuntimeError("Error in WIFISELECTIONCHANGED telemetry handler", err, frame.data)
				return
			}
			var	channel uint8
			binary.Read(bytes.NewReader(frame.data[12:13]), binary.LittleEndian, &channel)
			payload, _ := json.Marshal(struct{type, wfband string; channel int}{type:wftypestr, band:wfbandstr, channel: int(channel)})
			go sendTelemetry("networksettingsstate", payload)
		}
	default:
		{
			go sendUnknownTelemetry("Unknown Network commandId: "+strconv.Itoa(int(commandId)), frame.data)
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
			// No args.
			go sendEmptyTelemetry("flattrim")
		}
	// Flying state changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_FLYINGSTATECHANGED:
		{
			var flyingstate int
			binary.Read(bytes.NewReader(frame.data[4:8]), binary.LittleEndian, &flyingstate)
			// These are kind of a big deal so send them as separate events, unlike other enums
			switch byte(flyingstate) {
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_LANDED:
				{
					go sendEmptyTelemetry("landed")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_TAKINGOFF:
				{
					go sendEmptyTelemetry("takingoff")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_HOVERING:
				{
					go sendEmptyTelemetry("hovering")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_FLYING:
				{
					go sendEmptyTelemetry("flying")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_LANDING:
				{
					go sendEmptyTelemetry("landing")
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_EMERGENCY:
				{
					go sendEmptyTelemetry("emergency")
				}
			}
		}
	// Alert State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALERTSTATECHANGED:
		{
			statestr, err := decodeEnum(frame.data[4:8], []string{"none", "cut_out", "critical_battery", "low_battery", "too_much_angle"})
			if err != nil {
				go sendRuntimeError("Error in ALERTSTATECHANGED telemetry handler", err, frame.data)
				return
			}
			payload, _ := json.Marshal(struct{ state string }{state: statestr})
			go sendTelemetry("alertstate", payload)
		}
	// Navigate Home State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_NAVIGATEHOMESTATECHANGED:
		{
		  statestr, err := decodeEnum(frame.data[4:8], []string{"available", "inProgress", "unavailable", "pending"})
			if err != nil {
				go sendRuntimeError("Error in NAVIGATEHOMESTATECHANGED telemetry handler", err, frame.data)
				return
			}
			reasonstr, err := decodeEnum(frame.data[8:12], []string{"userRequest", "connectionLost", "lowBattery", "finished", "stopped", "disabled", "enabled"})
			if err != nil {
				go sendRuntimeError("Error in NAVIGATEHOMESTATECHANGED telemetry handler", err, frame.data)
				return
			}
			payload, _ := json.Marshal(struct{ state, reason string }{state: statestr, reason: reasonstr})
			go sendTelemetry("navigatehomestate", payload)
		}
	// Position (GPS)
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_POSITIONCHANGED:
		{
		  var telemdata struct{Lat, Lon, Alt float64}
			binary.Read(bytes.NewReader(frame.data[4:28]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go sendTelemetry("gps", payload)
		}
	// Speed Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_SPEEDCHANGED:
		{
			var telemdata struct{speedX, speedY, speedZ float64}
			binary.Read(bytes.NewReader(frame.data[4:28]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go sendTelemetry("speed", payload)
		}
	// Attitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ATTITUDECHANGED:
		{
			var telemdata struct{roll, pitch, yaw float32}
			binary.Read(bytes.NewReader(frame.data[4:16]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go sendTelemetry("attitude", payload)
		}
	// Auto Takeoff Mode Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_AUTOTAKEOFFMODECHANGED:
		{
			var telemdata struct{state bool}
			binary.Read(bytes.NewReader(frame.data[4:5]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go sendTelemetry("autotakeoffmode", payload)
		}
	// Altitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALTITUDECHANGED:
		{
			var telemdata struct{altitude float64}
			binary.Read(bytes.NewReader(frame.data[4:12]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go sendTelemetry("altitude", payload)
		}
	// End of PilotingState cases
	default:
		{
		  go sendUnknownTelemetry("Unknown Piloting State", frame.data)
		}
	}
}

func (b *Bebop) handleCommonStateFrame(commandId byte, frame *NetworkFrame) {
	switch commandId {
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_BATTERYSTATECHANGED:
		{
			// This uint8 is a percentage acc. to docs, should be 0-100?
			var telemdata struct{battery uint8}
			binary.Read(bytes.NewReader(frame.data[4:5]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("battery", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGESTATELISTCHANGED:
		{
			var (
				mass_storage_id uint8
				mass_stoage_name string
			)
			binary.Read(bytes.NewReader(frame.data[4:5]), binary.LittleEndian, &mass_storage_id)
			mass_storage_name = string(frame.data[5:])  // ? Encoding? Length? Huh?
			payload, _ := json.Marshal(struct{mass_storage_id uint8; name string}{mass_storage_id: mass_storage_id, name: mass_storage_name})
			go b.sendTelemetry("massstorage", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGEINFOSTATELISTCHANGED:
		// Information on a particular volume? Volunteered, or in response to a query?
		{
			var telemdata struct{mass_storage_id uint8; size, used_size uint32; plugged, full, internal bool}
			binary.Read(bytes.NewReader(frame.data[4:72]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("massstorageinfo", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTDATECHANGED:
		// Date in ISO-8601
		{
			dates := string(frame.data[4:])  // Parse to real time object? ISO-8601
			payload, _ := json.Marshal(struct{date string}{date: dates})
			go b.sendTelemetry("currentdate", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTTIMECHANGED:
		// Time in ISO-8601
		{
			times := string(frame.data[4:])  // Parse to real time object? ISO-8601
			payload, _ := json.Marshal(struct{time string}{time: times})
			go b.sendTelemetry("currenttime", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGEINFOREMAININGLISTCHANGED:
		// Remaining space on volume, with estimate of photo space/recording time?
		{
			var telemdata struct{free_space uint32; rec_time uint16; photo_remaining uint32}
			binary.Read(bytes.NewReader(frame.data[4:80]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("massstorageinforemaining", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_WIFISIGNALCHANGED:
		{
			var telemdata struct{rssi int16}  // in dbm
			binary.Read(bytes.NewReader(frame.data[4:20]), binary.LittleEndian, &telemdata)
			payload, _ := json.Marshal(telemdata)
			go b.sendTelemetry("wifisignal", payload)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_SENSORSSTATESLISTCHANGED:
		{
			var sensorState bool
			sensorName, err := decodeEnum(frame.data[4:8], []string{"IMU", "barometer", "ultrasound", "GPS", "magnetometer", "vertical_camera"})
			if err != nil {
				go sendRuntimeError("Error processing sensor state telemetry", err, frame)
				return
			}
			payload, _ := json.Marshal(struct{sensorName string; sensorState bool}{sensorName: sensorName, sensorState: sensorState})
			go b.sendTelemetry("sensorstates", payload)
		}
	default:
		{
			go sendUnknownTelemetry("Unknown/Unhandled COMMONSTATE commandId: "+strconv.Itoa(int(commandId)), frame.data)
		}
	}
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
	binary.Read(bytes.NewReader(frame.data[0:1]), binary.LittleEndian, &commandProject)
	binary.Read(bytes.NewReader(frame.data[1:2]), binary.LittleEndian, &commandClass)
	binary.Read(bytes.NewReader(frame.data[2:4]), binary.LittleEndian, &commandId)
	commandId = byte(commandId16)

	switch commandProject {
	case ARCOMMANDS_ID_PROJECT_COMMON:
		{
			switch commandClass {
			case ARCOMMANDS_ID_COMMON_CLASS_COMMONSTATE:
				{
					b.handleCommonStateFrame(commandId, frame)
				}
			default:
				{
					go sendUnknownTelemetry("Unknown/Unhandled common project commandClass: "+strconv.Itoa(int(commandClass)), frame.data)
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
			default:
				{
					go sendUnknownTelemetry("Unknown/Unhandled ARDRONE3 command Class: "+strconv.Itoa(int(commandClass)), frame.data)
				}
			}
		}
	default:
		{
			// This shouldn't happen, as there are only two expected projects?
			// Post an unknown telemetry event, may help to discover stuff for future usage.
			go sendUnknownTelemetry("Unknown Project: "+strconv.Itoa(int(commandProject)), frame.data)
		}
	}
}
