package client

import (
  "bytes"
  "strconv"
  "encoding/binary"
  "encoding/json"
)

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

func (b *Bebop) handleGPSSettingsStateFrame(commandId byte, frame *NetworkFrame) {
  switch commandId {
  case 0, 1: // HomeChanged - Return home status
             // ResetHomeChanged - Reset home status
    {
      var telemdata struct{
        Latitude float64 `json:"latitude"`
        Longitude float64 `json:"longitude"`
        Altitude float64 `json:"altitude"`
      }
      binary.Read(bytes.NewReader(frame.Data[4:28]), binary.LittleEndian, &telemdata)
      payload, _ := json.Marshal(telemdata)
      switch commandId {
      case 0:
        {
          go b.sendTelemetry("sethomechanged", payload)
        }
      case 1:
        {
          go b.sendTelemetry("resethomechanged", payload)
        }
      }
    }
  case 2: // GPSFixStateChanged - GPS fix state
    {
      var telemdata struct{
        Fixed bool `json:"fixed"`
      }
      binary.Read(bytes.NewReader(frame.Data[4:5]), binary.LittleEndian, &telemdata)
      payload, _ := json.Marshal(telemdata)
      go b.sendTelemetry("gpsfixstatechanged", payload)
    }
  case 3: // GPSUpdateStateChanged - GPS Update state
    {
      state, err := decodeEnum(frame.Data[4:8], []string{"updated", "inProgress", "failed"})
      if err != nil {
        b.sendRuntimeError("Failed to decode enum in GPSUpdateStateChanged telemetry", err, frame.Data)
        return
      }
      telemdata := struct{
        State string `json:"state"`
      }{state}
      payload, _ := json.Marshal(telemdata)
      go b.sendTelemetry("gpsupdatestatechanged", payload)
    }
  case 4: // HomeTypeChanged - State of the type of the home position. This type is the user preference. The prefered home type may not be available, see HomeTypeStatesChanged to get the drone home type
    {
      state, err := decodeEnum(frame.Data[4:8], []string{"TAKEOFF", "PILOT"})
      if err != nil {
        b.sendRuntimeError("Failed to decode enum in HomeTypeChanged telemetry", err, frame.Data)
        return
      }
      telemdata := struct{
        Type string `json:"type"`
      }{state}
      payload, _ := json.Marshal(telemdata)
      go b.sendTelemetry("hometypechanged", payload)
    }
  case 5: // ReturnHomeDelayChanged - State of the delay after which the drone will automatically try to return home
    {
      var telemdata struct{
        Delay uint16 `json:"delay"`
      }
      binary.Read(bytes.NewReader(frame.Data[4:6]), binary.LittleEndian, &telemdata)
      payload, _ := json.Marshal(telemdata)
      go b.sendTelemetry("returnhomedelaychanged", payload)
    }
  default:
    {
      go b.sendUnknownTelemetry("Unknown commandId in GPSSettingsState frame: "+strconv.Itoa(int(commandId)), frame.Data)
    }
  }
}
