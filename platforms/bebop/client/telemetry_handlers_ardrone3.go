package client

import (
	"bytes"
	"encoding/binary"
)

// Handles the important events that related to device state in the air: GPS position,
// attitude, speed, etcetera.
func (b *Bebop) handlePilotingStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	// Flat Trim changed (?)
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_FLATTRIMCHANGED:
		{
			// No args. Very often.
			err = b.sendEmptyTelemetry("flattrim")
			if err != nil {
				return true, "FlatTrimChanged", err
			}
		}
	// Flying state changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_FLYINGSTATECHANGED:
		{
			var flyingstate int
			err = binary.Read(bytes.NewReader(frame.Data[4:8]), binary.LittleEndian, &flyingstate)
			if err != nil {
				return true, "FlyingStateChanged", err
			}
			// These are kind of a big deal so send them as separate events, unlike other enums
			switch byte(flyingstate) {
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_LANDED:
				{
					err = b.sendEmptyTelemetry("landed")
					if err != nil {
						return true, "FlyingStateChanged:Landed", err
					}
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_TAKINGOFF:
				{
					err = b.sendEmptyTelemetry("takingoff")
					if err != nil {
						return true, "FlyingStateChanged:TakingOff", err
					}
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_HOVERING:
				{
					err = b.sendEmptyTelemetry("hovering")
					if err != nil {
						return true, "FlyingStateChanged:Hovering", err
					}
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_FLYING:
				{
					err = b.sendEmptyTelemetry("flying")
					if err != nil {
						return true, "FlyingStateChanged:Flying", err
					}

				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_LANDING:
				{
					err = b.sendEmptyTelemetry("landing")
					if err != nil {
						return true, "FlyingStateChanged:Landing", err
					}
				}
			case ARCOMMANDS_ARDRONE3_PILOTINGSTATE_FLYINGSTATECHANGED_STATE_EMERGENCY:
				{
					err = b.sendEmptyTelemetry("emergency")
					if err != nil {
						return true, "FlyingStateChanged:Emergency", err
					}
				}
			default:
				{
					// If this does occur it'll take a minute to figure out, but worth
					// covering that base.
					return false, "", nil
				}
			}
		}
	// Alert State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALERTSTATECHANGED:
		{
			statestr, err := decodeEnum(frame.Data[4:8], []string{"none", "cut_out", "critical_battery", "low_battery", "too_much_angle"})
			if err != nil {
				return true, "AlertStateChanged", err
			}
			err = b.sendJSONTelemetry(frame, "alertstate", struct {
				State string `json:"state"`
			}{State: statestr})
			if err != nil {
				return true, "AlertStateChanged", err
			}
		}
	// Navigate Home State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_NAVIGATEHOMESTATECHANGED:
		{
			statestr, err := decodeEnum(frame.Data[4:8], []string{"available", "inProgress", "unavailable", "pending"})
			if err != nil {
				return true, "NavigateHomeStateChanged", err
			}
			reasonstr, err := decodeEnum(frame.Data[8:12], []string{"userRequest", "connectionLost", "lowBattery", "finished", "stopped", "disabled", "enabled"})
			if err != nil {
				return true, "NavigateHomeStateChanged", err
			}
			err = b.sendJSONTelemetry(frame, "navigatehomestate", struct {
				State  string `json:"state"`
				Reason string `json:"reason"`
			}{
				State: statestr, Reason: reasonstr,
			})
			if err != nil {
				return true, "NavigateHomeStateChanged", err
			}
		}
	// Position (GPS)
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_POSITIONCHANGED:
		{
			var telemdata struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
				Alt float64 `json:"alt"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:28]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "PositionChanged", err
			}
			err = b.sendJSONTelemetry(frame, "gps", telemdata)
			if err != nil {
				return true, "PositionChanged", err
			}
		}
	// Speed Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_SPEEDCHANGED:
		{
			var telemdata struct {
				SpeedX float64 `json:"speedX"`
				SpeedY float64 `json:"speedY"`
				SpeedZ float64 `json:"speedZ"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:28]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "SpeedChanged", err
			}
			err = b.sendJSONTelemetry(frame, "speed", telemdata)
			if err != nil {
				return true, "SpeedChanged", err
			}
		}
	// Attitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ATTITUDECHANGED:
		{
			var telemdata struct {
				Roll  float32 `json:"roll"`
				Pitch float32 `json:"pitch"`
				Yaw   float32 `json:"yaw"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "AttitudeChanged", err
			}
			err = b.sendJSONTelemetry(frame, "attitude", telemdata)
			if err != nil {
				return true, "AttitudeChanged", err
			}
		}
	// Auto Takeoff Mode Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_AUTOTAKEOFFMODECHANGED:
		{
			var telemdata struct {
				State bool `json:"state"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "AutoTakeoffModeChanged", err
			}
			err = b.sendJSONTelemetry(frame, "autotakeoffmode", telemdata)
			if err != nil {
				return true, "AutoTakeoffModeChanged", err
			}
		}
	// Altitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALTITUDECHANGED:
		{
			var telemdata struct {
				Altitude float64 `json:"altitude"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:12]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "AltitudeChanged", err
			}
			err = b.sendJSONTelemetry(frame, "altitude", telemdata)
			if err != nil {
				return true, "AltitudeChanged", err
			}
		}
	// End of PilotingState cases
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

func (b *Bebop) handleGPSSettingsStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0, 1: // HomeChanged - Return home status
		// ResetHomeChanged - Reset home status
		{
			var (
				telemdata struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
					Altitude  float64 `json:"altitude"`
				}
				EventName, EventTitle string
			)
			switch commandId {
			case 0:
				{
					EventName = "SetHomeChanged"
					EventTitle = "sethomechanged"
				}
			case 1:
				{
					EventName = "ResetHomeChanged"
					EventTitle = "resethomechanged"
				}
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:28]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, EventName, err
			}
			err = b.sendJSONTelemetry(frame, EventTitle, telemdata)
			if err != nil {
				return true, EventName, err
			}
		}
	case 2: // GPSFixStateChanged - GPS fix state
		{
			var telemdata struct {
				Fixed bool `json:"fixed"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "GPSFixStateChanged", err
			}
			err = b.sendJSONTelemetry(frame, "gpsfixstatechanged", telemdata)
			if err != nil {
				return true, "GPSFixStateChanged", err
			}
		}
	case 3: // GPSUpdateStateChanged - GPS Update state
		{
			state, err := decodeEnum(frame.Data[4:8], []string{"updated", "inProgress", "failed"})
			if err != nil {
				return true, "GPSUpdateStateChanged", err
			}
			telemdata := struct {
				State string `json:"state"`
			}{state}
			err = b.sendJSONTelemetry(frame, "gpsupdatestatechanged", telemdata)
			if err != nil {
				return true, "GPSUpdateStateChanged", err
			}
		}
	case 4: // HomeTypeChanged - State of the type of the home position. This type is the user preference. The prefered home type may not be available, see HomeTypeStatesChanged to get the drone home type
		{
			state, err := decodeEnum(frame.Data[4:8], []string{"TAKEOFF", "PILOT"})
			if err != nil {
				return true, "HomeTypeChanged", err
			}
			telemdata := struct {
				Type string `json:"type"`
			}{state}
			err = b.sendJSONTelemetry(frame, "hometypechanged", telemdata)
			if err != nil {
				return true, "HomeTypeChanged", err
			}
		}
	case 5: // ReturnHomeDelayChanged - State of the delay after which the drone will automatically try to return home
		{
			var telemdata struct {
				Delay uint16 `json:"delay"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:6]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "ReturnHomeDelayChanged", err
			}
			err = b.sendJSONTelemetry(frame, "returnhomedelaychanged", telemdata)
			if err != nil {
				return true, "ReturnHomeDelayChanged", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

func (b *Bebop) handleCameraStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0:
		{
			// Only one commandId, 0. Don't bother checking?
			var telemdata struct{ Tilt, Pan int8 }
			err = binary.Read(bytes.NewReader(frame.Data[4:6]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "CameraState", err
			}
			err = b.sendJSONTelemetry(frame, "camerastate", telemdata)
			if err != nil {
				return true, "CameraState", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

// Handle telemetry from device pertaining to Wifi band/channel settings
func (b *Bebop) handleNetworkSettingsStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case ARCOMMANDS_ARDRONE3_NETWORKSETTINGSSTATECHANGED_STATE_WIFISELECTIONCHANGED:
		// Appears to be simply feedback for when the client issues a corresponding
		// instruction; returns settings to confirm?
		{
			var channel uint8
			wftypestr, err := decodeEnum(frame.Data[4:8], []string{"auto_all", "auto_2_4ghz", "auto_5ghz", "all"})
			if err != nil {
				return true, "WifiSelectionChanged", err
			}
			wfbandstr, err := decodeEnum(frame.Data[8:12], []string{"2_4ghz", "5ghz", "all"})
			if err != nil {
				return true, "WifiSelectionChanged", err
			}
			err = binary.Read(bytes.NewReader(frame.Data[12:13]), binary.LittleEndian, &channel)
			if err != nil {
				return true, "WifiSelectionChanged", err
			}
			err = b.sendJSONTelemetry(frame, "networksettingsstate", struct {
				Type    string `json:"type"`
				Band    string `json:"band"`
				Channel int    `json:"channel"`
			}{Type: wftypestr, Band: wfbandstr, Channel: int(channel)})
			if err != nil {
				return true, "WifiSelectionChanged", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

// Handles events about the camera. These seem to mostly be confirmation of user-set
// camera parameters.
func (b *Bebop) handlePictureSettingsStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_PICTUREFORMATCHANGED:
		{
			types, err := decodeEnum(frame.Data[4:8], []string{"raw", "jpeg", "snapshot"})
			if err != nil {
				return true, "PictureFormatChanged", err
			}
			err = b.sendJSONTelemetry(frame, "pictureformatchanged", struct {
				Type string `json:"type"`
			}{Type: types})
			if err != nil {
				return true, "PictureFormatChanged", err
			}
		}
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_AUTOWHITEBALANCECHANGED:
		{
			types, err := decodeEnum(frame.Data[4:8], []string{"auto", "tungsten", "daylight", "cloudy", "cool_white"})
			if err != nil {
				return true, "AutoWhiteBalanceChanged", err
			}
			err = b.sendJSONTelemetry(frame, "autowhitebalancechanged", struct {
				Type string `json:type`
			}{Type: types})
			if err != nil {
				return true, "AutoWhiteBalanceChanged", err
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
			err = binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "ExpositionChanged/SaturationChanged", err
			}
			switch commandId {
			case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_EXPOSITIONCHANGED:
				{
					err = b.sendJSONTelemetry(frame, "expositionchanged", telemdata)
					if err != nil {
						return true, "ExpositionChanged", err
					}
				}
			case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_SATURATIONCHANGED:
				{
					err = b.sendJSONTelemetry(frame, "saturationchanged", telemdata)
					if err != nil {
						return true, "SaturationChanged", err
					}
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
			err = binary.Read(bytes.NewReader(frame.Data[4:17]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "TimeLapseChanged", err
			}
			err = b.sendJSONTelemetry(frame, "timelapsechanged", telemdata)
			if err != nil {
				return true, "TimeLapseChanged", err
			}
		}
	case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_VIDEOAUTORECORDCHANGED:
		{
			var telemdata struct {
				Enabled         bool  `json:"enabled"`
				Mass_storage_id uint8 `json:"mass_storage_id"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:6]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "VideoAutoRecordChanged", err
			}
			err = b.sendJSONTelemetry(frame, "videoautorecordchanged", telemdata)
			if err != nil {
				return true, "VideoAutoRecordChanged", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}
