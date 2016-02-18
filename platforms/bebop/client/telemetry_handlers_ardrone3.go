package client

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/hybridgroup/gobot/platforms/bebop/bbtelem"
)

// The "Camera" set are commands, not telemetry, so unsure why it's sending with
// this ID. Let's implement and see?
func (b *Bebop) handleCameraFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0: // Orientation; this is an instruction *to* the drone. It shouldn't even be here.
		{
			var telemdata struct {
				Tilt uint8 `json:"tilt"`
				Pan  uint8 `json:"pan"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:6]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "Orientation", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Orientation, telemdata)
			if err != nil {
				return true, "Orientation", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

// Handles the important events that related to device state in the air: GPS position,
// attitude, speed, etcetera.
func (b *Bebop) handlePilotingStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	// Flat Trim changed (?) == 0
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_FLATTRIMCHANGED:
		{
			// No args. Very often.
			err = b.sendEmptyTelemetry(bbtelem.Flattrim)
			if err != nil {
				return true, "FlatTrimChanged", err
			}
		}
	// Flying state changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_FLYINGSTATECHANGED:
		{
			/*states := []string{"landed",
			"takingoff",
			"hovering",
			"flying",
			"landing",
			"emergency"}  //*/
			states := []string{
				bbtelem.Landed,    // "bebop:landed"
				bbtelem.Takingoff, // "bebop:takingoff"
				bbtelem.Hovering,  // "bebop:hovering"
				bbtelem.Flying,    // "bebop:flying"
				bbtelem.Landing,   // "bebop:landing"
				bbtelem.Emergency, // "bebop:emergency"
			}
			flyingstate, err := decodeEnum(frame.Data[4:8], states)
			if err != nil {
				return true, "FlyingStateChanged", err
			}
			err = b.sendEmptyTelemetry(flyingstate)
			if err != nil {
				return true, "FlyingStateChanged", err
			}
		}
	// Alert State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALERTSTATECHANGED:
		{
			statestr, err := decodeEnum(frame.Data[4:8], []string{"none", "user", "cut_out", "critical_battery", "low_battery", "too_much_angle"})
			if err != nil {
				return true, "AlertStateChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Alertstate, struct {
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
			err = b.sendJSONTelemetry(frame, bbtelem.Navigatehomestate, struct {
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
			err = b.sendJSONTelemetry(frame, bbtelem.Gps, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Speed, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Attitude, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Autotakeoffmode, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Altitude, telemdata)
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

var notImplementedInFirmwareError = errors.New("Not implemented in Firmware yet, presumed impossible?")

// Reports a bunch of settings pertaining to maxima, minima, and boolean switches
// like "obey max height". Lots of unimplemented draft settings, too.
func (b *Bebop) handlePilotingSettingsState(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0:
		{
			// MaxAltitudeChanged
			var telemdata struct {
				Current float32 `json:"current"`
				Min     float32 `json:"min"`
				Max     float32 `json:"max"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "MaxAltitudeChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Maxaltitudechanged, telemdata)
			if err != nil {
				return true, "MaxAltitudeChanged", err
			}
		}
	case 1:
		{
			// MaxTiltChanged
			var telemdata struct {
				Current float32 `json:"current"`
				Min     float32 `json:"min"`
				Max     float32 `json:"max"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "MaxTiltChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Maxtiltchanged, telemdata)
			if err != nil {
				return true, "MaxTiltChanged", err
			}
		}
	case 2:
		{
			// AbsolutControlChanged
			var telemdata struct {
				On uint8 `json:"on"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "AbsolutControlChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Absolutcontrolchanged, telemdata)
			if err != nil {
				return true, "AbsolutControlChanged", err
			}
		}
	case 3:
		{
			// MaxDistanceChanged
			var telemdata struct {
				Current float32 `json:"current"`
				Min     float32 `json:"min"`
				Max     float32 `json:"max"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "MaxDistanceChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Maxdistancechanged, telemdata)
			if err != nil {
				return true, "MaxDistanceChanged", err
			}
		}
	// 4 = NoFlyOverMaxDistanceChanged
	case 4:
		{
			// NoFlyOverMaxDistanceChanged
			var telemdata struct {
				ShouldNotFlyOver uint8 `json:"shouldNotFlyOver"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "NoFlyOverMaxDistanceChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Noflyovermaxdistancechanged, telemdata)
			if err != nil {
				return true, "NoFlyOverMaxDistanceChanged", err
			}
		}
	case 5: // AutonomousFlightMaxHorizontalSpeed (Not Implemented in Firmware) (But is it numerically assigned??)
		{
			return true, "AutonomousFlightMaxHorizontalSpeed", notImplementedInFirmwareError
		}
	case 6: // AutonomousFlightMaxVerticalSpeed (Not Implemented in Firmware) (But is it numerically assigned??)
		{
			return true, "AutonomousFlightMaxVerticalSpeed", notImplementedInFirmwareError
		}
	case 7: // AutonomousFlightMaxHorizontalAcceleration (Not Implemented in Firmware)
		{
			return true, "AutonomousFlightMaxHorizontalAcceleration", notImplementedInFirmwareError
		}
	case 8: // AutonomousFlightMaxVerticalAcceleration (Not implemented in firmware)
		{
			return true, "AutonomousFlightMaxVerticalAcceleration", notImplementedInFirmwareError
		}
	case 9: // AutonomousFlightMaxRotationSpeed (Not implemented in firmware)
		{
			return true, "AutonomousFlightMaxRotationSpeed", notImplementedInFirmwareError
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}

func (b *Bebop) handleSpeedSettingsState(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0:
		// MaxVerticalSpeedChanged - Max vertical speed sent by product
		{
			var telemdata struct {
				Current float32 `json:"current"`
				Min     float32 `json:"min"`
				Max     float32 `json:"max"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "MaxVerticalSpeedChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Maxverticalspeedchanged, telemdata)
			if err != nil {
				return true, "MaxVerticalSpeedChanged", err
			}
		}
	case 1:
		// MaxRotationSpeedChanged - Max rotation speed sent by product
		{
			var telemdata struct {
				Current float32 `json:"current"`
				Min     float32 `json:"min"`
				Max     float32 `json:"max"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:16]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "MaxRotationSpeedChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Maxrotationspeedchanged, telemdata)
			if err != nil {
				return true, "MaxRotationSpeedChanged", err
			}
		}
	case 2:
		// HullProtectionChanged - Presence of hull protection sent by product
		{
			var telemdata struct {
				Present uint8 `json:"present"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "HullProtectionChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Hullprotectionchanged, telemdata)
			if err != nil {
				return true, "HullProtectionChanged", err
			}
		}
	case 3:
		// OutdoorChanged - Outdoor property sent by product
		{
			var telemdata struct {
				Present uint8 `json:"present"`
			}
			err = binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
			if err != nil {
				return true, "OutdoorChanged", err
			}
			err = b.sendJSONTelemetry(frame, bbtelem.Outdoorchanged, telemdata)
			if err != nil {
				return true, "OutdoorChanged", err
			}
		}
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
					EventTitle = bbtelem.Sethomechanged
				}
			case 1:
				{
					EventName = "ResetHomeChanged"
					EventTitle = bbtelem.Resethomechanged
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
			err = b.sendJSONTelemetry(frame, bbtelem.Gpsfixstatechanged, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Gpsupdatestatechanged, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Hometypechanged, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Returnhomedelaychanged, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Camerastate, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Networksettingsstate, struct {
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
			err = b.sendJSONTelemetry(frame, bbtelem.Pictureformatchanged, struct {
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
			err = b.sendJSONTelemetry(frame, bbtelem.Autowhitebalancechanged, struct {
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
					err = b.sendJSONTelemetry(frame, bbtelem.Expositionchanged, telemdata)
					if err != nil {
						return true, "ExpositionChanged", err
					}
				}
			case ARCOMMANDS_ARDRONE3_PICTURESETTINGSSTATECHANGED_STATE_SATURATIONCHANGED:
				{
					err = b.sendJSONTelemetry(frame, bbtelem.Saturationchanged, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Timelapsechanged, telemdata)
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
			err = b.sendJSONTelemetry(frame, bbtelem.Videoautorecordchanged, telemdata)
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

func (b *Bebop) handleNetworkStateFrame(commandId byte, frame *NetworkFrame) (found bool, context string, err error) {
	switch commandId {
	case 0: //WifiScanListChanged (This is a Map, followed by AllWifiScanChanged to signal end?)
		{
			var (
				Ssid    string
				Rssi    int16
				Band    string
				Channel uint8
				rest    []byte
			)
			Ssid, rest, err = parseNullTermedString(frame.Data[4:])
			if err != nil {
				return true, "WifiScanListChanged", err
			}
			Rssi = int16(binary.LittleEndian.Uint16(rest[:2]))
			Band, err = decodeEnum(rest[2:6], []string{"2_4ghz", "5ghz"})
			if err != nil {
				return true, "WifiScanListChanged", err
			}
			Channel = uint8(rest[6])
			telemdata := struct {
				Ssid    string `json:"ssid"`
				Rssi    int16  `json:"rssi"`
				Band    string `json:"band"`
				Channel uint8  `json:"channel"`
			}{Ssid, Rssi, Band, Channel}
			err = b.sendJSONTelemetry(frame, bbtelem.Wifiscanlistchanged, telemdata)
			if err != nil {
				return true, "WifiScanListChanged", err
			}
		}
	case 1: // AllWifiScanChanged - Sent when WifiScanListChanged events are finished?
		{
			err = b.sendEmptyTelemetry(bbtelem.Allwifiscanchanged)
			if err != nil {
				return true, "AllWifiScanChanged", err
			}
		}
	case 2: // WifiAuthChannelListChanged, indicates which channels are permitted for use outside?
		// Is this a geography/legislation thing? "List" suggests order is important, seems not though?
		{
			var (
				Band           string
				Channel        uint8
				InOrOut        uint8
				AllowedOutside bool
				AllowedInside  bool
			)
			Band, err = decodeEnum(frame.Data[4:8], []string{"2_4ghz", "5ghz"})
			if err != nil {
				return true, "WifiAuthChannelListChanged", err
			}
			Channel = uint8(frame.Data[8])
			InOrOut = uint8(frame.Data[9])
			AllowedOutside = 0 < (InOrOut & 1)
			AllowedInside = 0 < (InOrOut & 2)
			telemdata := struct {
				Band           string `json:"band"`
				Channel        uint8  `json:"channel"`
				InOrOut        uint8  `json:"in_or_out"`
				AllowedOutside bool   `json:"allowedOutside"`
				AllowedInside  bool   `json:"allowedInside"`
			}{Band, Channel, InOrOut, AllowedOutside, AllowedInside}
			err = b.sendJSONTelemetry(frame, bbtelem.Wifiauthchannellistchanged, telemdata)
			if err != nil {
				return true, "WifiAuthChannelListChanged", err
			}
		}
	case 3:
		{
			// AllWifiAuthChannelChanged, sent when authorised list is fully sent.
			err = b.sendEmptyTelemetry(bbtelem.Allwifiauthchannelchanged)
			if err != nil {
				return true, "AllWifiAuthChannelChanged", err
			}
		}
	default:
		{
			return false, "", nil
		}
	}
	return true, "", nil
}
