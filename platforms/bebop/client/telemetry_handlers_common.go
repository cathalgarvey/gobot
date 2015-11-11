package client

import (
	"bytes"
	"encoding/binary"
)

// Internal states, like settings, battery level, storage, date/time,
func (b *Bebop) handleCommonStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_ALLSTATESCHANGED:
		{
			// Is this even useful telemetry? Ignoring for now
			err = b.sendEmptyTelemetry("allstateschanged")
			if err != nil {
				return true, "AllStatesChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_BATTERYSTATECHANGED:
		{
			// This uint8 is a percentage acc. to docs, should be 0-100?
			var telemdata struct {
				Battery uint8 `json:"battery"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "BatteryStateChanged", err
			}
			err = b.sendJSONTelemetry(frame, "battery", telemdata)
			if err != nil {
				return true, "BatteryStateChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGESTATELISTCHANGED:
		{
			var (
				mass_storage_id uint8
			)
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &mass_storage_id)
			if err != nil {
				return true, "MassStorageStateListChanged", err
			}
			mass_storage_name := string(frame.Data[5:]) // ? Encoding? Length? Huh?
			err = b.sendJSONTelemetry(frame, "massstorage", struct {
				Mass_storage_id uint8  `json:"Mass_storage_id"`
				Name            string `json:"name"`
			}{Mass_storage_id: mass_storage_id, Name: mass_storage_name})
			if err != nil {
				return true, "MassStorageStateListChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGEINFOSTATELISTCHANGED:
		// Information on a particular volume? Volunteered, or in response to a query?
		{
			var telemdata struct {
				Mass_storage_id uint8  `json:"mass_storage_id"`
				Size            uint32 `json:"size"`
				Used_size       uint32 `json:"used_size"`
				Plugged         uint8   `json:"plugged"`
				Full            uint8   `json:"full"`
				Internal        uint8   `json:"internal"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:72]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "MassStorageInfoStateListChanged", err
			}
			err = b.sendJSONTelemetry(frame, "massstorageinfo", telemdata)
			if err != nil {
				return true, "MassStorageInfoStateListChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTDATECHANGED:
		// Date in ISO-8601
		{
			dates, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				return true, "CurrentDateChanged", err
			}
			err = b.sendJSONTelemetry(frame, "currentdate", struct {
				Date string `json:"date"`
			}{Date: dates})
			if err != nil {
				return true, "CurrentDateChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTTIMECHANGED:
		// Time in ISO-8601
		{
			times, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				return true, "CurrentTimeChanged", err
			}
			err = b.sendJSONTelemetry(frame, "currenttime", struct {
				Time string `json:"time"`
			}{Time: times})
			if err != nil {
				return true, "CurrentTimeChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_MASSSTORAGEINFOREMAININGLISTCHANGED:
		// Remaining space on volume, with estimate of photo space/recording time?
		{
			var telemdata struct {
				Free_space      uint32 `json:"free_space"`
				Rec_time        uint16 `json:"rec_time"`
				Photo_remaining uint32 `json:"photo_remaining"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:80]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "MassStorageInfoRemainingListChanged", err
			}
			err = b.sendJSONTelemetry(frame, "massstorageinforemaining", telemdata)
			if err != nil {
				return true, "MassStorageInfoRemainingListChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_WIFISIGNALCHANGED:
		{
			var telemdata struct {
				Rssi int16 `json:"rssi"`
			} // in dbm
			err = binary.Read(bytes.NewReader(frame.Data[4:20]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "WifiSignalChanged", err
			}
			err = b.sendJSONTelemetry(frame, "wifisignal", telemdata)
			if err != nil {
				return true, "WifiSignalChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_SENSORSSTATESLISTCHANGED:
		{
			var sensorState bool
			sensorName, err := decodeEnum(frame.Data[4:8], []string{"IMU", "barometer", "ultrasound", "GPS", "magnetometer", "vertical_camera"})
			if err != nil {
				return true, "SensorStatesListChanged", err
			}
			err = b.sendJSONTelemetry(frame, "sensorstates", struct {
				SensorName  string `json:"sensorName"`
				SensorState bool   `json:"sensorState"`
			}{SensorName: sensorName, SensorState: sensorState})
			if err != nil {
				return true, "SensorStatesListChanged", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_PRODUCTMODEL:
		// This appears to be irrelevant to the Bebop but it's in "common"!
		{
			modelstr, err := decodeEnum(frame.Data[4:8], []string{"RS_TRAVIS", "RS_MARS", "RS_SWAT", "RS_MCLANE", "RS_BLAZE", "RS_ORAK", "RS_NEWZ", "JS_DIESEL", "JS_BUZZ", "JS_MAX", "JS_JETT", "JS_TUKTUK"})
			if err != nil {
				return true, "ProductModel", err
			}
			err = b.sendJSONTelemetry(frame, "dronemodel", struct {
				Model string `json:"model"`
			}{Model: modelstr})
			if err != nil {
				return true, "ProductModel", err
			}
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_COUNTRYLISTKNOWN:
		{
			ccodes := string(frame.Data[4:])
			err = b.sendJSONTelemetry(frame, "countrycodes", struct {
				CountryCodes string `json:"countryCodes"`
			}{ccodes})
			if err != nil {
				return true, "CountryListKnown", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

// Device can volunteer version info sometimes.
func (b *Bebop) handleVersionStateFrames(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0: // ControllerLibARCommandsVersion
		{
			version, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				return true, "ControllerLibARCommandsVersion", err
			}
			err = b.sendJSONTelemetry(frame, "controllerlibversion", struct{ Version string }{Version: version})
			if err != nil {
				return true, "ControllerLibARCommandsVersion", err
			}
		}
	case 1: // SkyControllerLibARCommandsVersion
		{
			version, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				return true, "SkycontrollerLibARCommandsVersion", err
			}
			err = b.sendJSONTelemetry(frame, "skycontrollerlibversion", struct{ Version string }{Version: version})
			if err != nil {
				return true, "SkycontrollerLibARCommandsVersion", err
			}
		}
	case 2: // DeviceLibARCommandsVersion
		{
			version, _, err := parseNullTermedString(frame.Data[4:])
			if err != nil {
				return true, "DeviceLibARCommandsVersion", err
			}
			err = b.sendJSONTelemetry(frame, "devicelibversion", struct{ Version string }{Version: version})
			if err != nil {
				return true, "DeviceLibARCommandsVersion", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

// Handle common Mavlink/Flightplan state frame
func (b *Bebop) handleMavlinkStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0: // MavlinkFilePlayingStateChanged,  Playing state of a mavlink flight plan
		{
			state, err := decodeEnum(frame.Data[4:8], []string{"playing", "stopped", "paused"})
			if err != nil {
				return true, "MavlinkFilePlayingStateChanged", err
			}
			filepath, rest, err := parseNullTermedString(frame.Data[8:])
			if err != nil {
				return true, "MavlinkFilePlayingStateChanged", err
			}
			types, err := decodeEnum(rest, []string{"flightPlan", "mapMyHouse"})
			if err != nil {
				return true, "MavlinkFilePlayingStateChanged", err
			}
			err = b.sendJSONTelemetry(frame, "mavlinkfileplaying", struct {
				State    string `json:"state"`
				Filepath string `json:"filepath"`
				Type     string `json:"type"`
			}{state, filepath, types})
			if err != nil {
				return true, "MavlinkFilePlayingStateChanged", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

func (b *Bebop) handleCameraSettingsState(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	// Appears in static log
	// Only one command, "camerasettingsstate", Id == 0
	if commandId != 0 {
		return false, "", nil
	}
	var telemdata struct {
		Fov     float32 `json:"fov"`
		PanMax  float32 `json:"panMax"`
		PanMin  float32 `json:"panMin"`
		TiltMax float32 `json:"tiltMax"`
		TileMin float32 `json:"tileMin"`
	}
	err = binary.Read(bytes.NewReader(frame.Data[4:4+(32*5)]), binary.LittleEndian, &telemdata)
	if err != nil {
		return true, "CameraSettingsState", err
	}
	err = b.sendJSONTelemetry(frame, "camerasettingsstate", telemdata)
	if err != nil {
		return true, "CameraSettingsState", err
	}
	return true, "", nil
}

func (b *Bebop) handleFlightPlanState(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	// Dumps regularly
	// One command "AvailabilityStateChanged", Id == 0
	if commandId != 0 {
		return false, "", nil
	}
	var telemdata struct {
		AvailabilityState uint8 `json:"availabilityState"`
	}
	err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
	if err != nil {
		return true, "AvailabilityStateChanged", err
	}
	err = b.sendJSONTelemetry(frame, "availabilitystatechanged", telemdata)
	if err != nil {
		return true, "AvailabilityStateChanged", err
	}
	return true, "", nil
}

func (b *Bebop) handleFlightPlanEvent(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0: // StartingErrorEvent - Event of flight plan start error
		{
			err = b.sendEmptyTelemetry("startingerrorevent")
			if err != nil {
				return true, "StartingErrorEvent", err
			}
		}
	case 1: // SpeedBridleEvent - Bridle speed of the drone
		{
			err = b.sendEmptyTelemetry("speedbridleevent")
			if err != nil {
				return true, "SpeedBridleEvent", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

// TODO: Handle! ARCOMMANDS_ID_COMMON_CLASS_SETTINGSSTATE
func (b *Bebop) handleEventCommonSettingsState(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	return false, "", nil
	// ARCOMMANDS_ID_COMMON_SETTINGSSTATE_CMD_ALLSETTINGSCHANGED = 0
	// ARCOMMANDS_ID_COMMON_SETTINGSSTATE_CMD_PRODUCTNAMECHANGED = 2
	// ARCOMMANDS_ID_COMMON_SETTINGSSTATE_CMD_PRODUCTVERSIONCHANGED = 3
	// ARCOMMANDS_ID_COMMON_SETTINGSSTATE_CMD_PRODUCTSERIALHIGHCHANGED = 4
	// ARCOMMANDS_ID_COMMON_SETTINGSSTATE_CMD_PRODUCTSERIALLOWCHANGED = 5
	// ARCOMMANDS_ID_COMMON_SETTINGSSTATE_CMD_COUNTRYCHANGED = 6
	// ARCOMMANDS_ID_COMMON_SETTINGSSTATE_CMD_AUTOCOUNTRYCHANGED = 7
}

func (b *Bebop) handleNetworkFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	// Single-command Class
	if commandId != 0 {
		return false, "", nil
	}
	err = b.sendEmptyTelemetry("networkdisconnect")
	if err != nil {
		return true, "NetworkDisconnect", err
	}
	return true, "", nil
}
