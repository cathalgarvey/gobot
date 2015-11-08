package client

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

// Shortcut method for sending a title with no data.
func (b *Bebop) sendEmptyTelemetry(title string) {
	sendTelementry(title, []byte{})
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
			var (
				state    int
				statestr string
			)
			binary.Read(bytes.NewReader(frame.data[4:8]), binary.LittleEndian, &state)
			statestr = []string{"none", "cut_out", "critical_battery", "low_battery", "too_much_angle"}[state]
			payload, _ := json.Marshal(struct{ state string }{state: statestr})
			go sendTelemetry("alertstate", payload)
		}
	// Navigate Home State Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_NAVIGATEHOMESTATECHANGED:
		{
			var (
				state, reason       int
				statestr, reasonstr string
			)
			binary.Read(bytes.NewReader(frame.data[4:8]), binary.LittleEndian, &state)
			binary.Read(bytes.NewReader(frame.data[8:12]), binary.LittleEndian, &reason)
			statestr = []string{"available", "inProgress", "unavailable", "pending"}[state]
			reasonstr = []string{"userRequest", "connectionLost", "lowBattery", "finished", "stopped", "disabled", "enabled"}[reason]
			payload, _ := json.Marshal(struct{ state, reason string }{state: statestr, reason: reasonstr})
			go sendTelemetry("navigatehomestate", payload)
		}
	// Position (GPS)
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_POSITIONCHANGED:
		{
			var (
				lat, lon, alt float64
			)
			binary.Read(bytes.NewReader(frame.data[4:12]), binary.LittleEndian, &lat)
			binary.Read(bytes.NewReader(frame.data[12:20]), binary.LittleEndian, &lon)
			binary.Read(bytes.NewReader(frame.data[20:28]), binary.LittleEndian, &alt)
			payload, _ := json.Marshal(struct{ Lat, Lon, Alt float64 }{Lat: lat, Lon: lon, Alt: alt})
			go sendTelemetry("gps", payload)
		}
	// Speed Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_SPEEDCHANGED:
		{
			var (
				speedX, speedY, speedZ float32
			)
			binary.Read(bytes.NewReader(frame.data[4:8]), binary.LittleEndian, &speedX)
			binary.Read(bytes.NewReader(frame.data[8:12]), binary.LittleEndian, &speedY)
			binary.Read(bytes.NewReader(frame.data[12:16]), binary.LittleEndian, &speedZ)
			payload, _ := json.Marshal(struct{ speedX, speedY, speedZ float32 }{speedX: speedX, speedY: speedY, speedZ: speedZ})
			go sendTelemetry("speed", payload)
		}
	// Attitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ATTITUDECHANGED:
		{
			var (
				roll, pitch, yaw float32
			)
			binary.Read(bytes.NewReader(frame.data[4:8]), binary.LittleEndian, &roll)
			binary.Read(bytes.NewReader(frame.data[8:12]), binary.LittleEndian, &pitch)
			binary.Read(bytes.NewReader(frame.data[12:16]), binary.LittleEndian, &yaw)
			payload, _ := json.Marshal(struct{ roll, pitch, yaw float32 }{roll: roll, pitch: pitch, yaw: yaw})
			go sendTelemetry("attitude", payload)
		}
	// Auto Takeoff Mode Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_AUTOTAKEOFFMODECHANGED:
		{
			var state uint8
			binary.Read(bytes.NewReader(frame.data[4:5]), binary.LittleEndian, &state)
			payload, _ := json.Marshal(struct{ state int }{state: int(state)})
			go sendTelemetry("autotakeoffmode", payload)
		}
	// Altitude Changed
	case ARCOMMANDS_ID_ARDRONE3_PILOTINGSTATE_CMD_ALTITUDECHANGED:
		{
			var altitude float64
			binary.Read(bytes.NewReader(frame.data[4:12]), binary.LittleEndian, &altitude)
			payload, _ := json.Marshal(struct{ altitude float64 }{altitude: altitude})
			go sendTelemetry("altitude", payload)
		}
	// End of PilotingState cases
	default:
		{
			go sendTelemetry("unknownPilotingState", frame.data)
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
					//
					switch commandId {
					case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_BATTERYSTATECHANGED:
						{
							battery_data := byte(0)
							binary.Read(bytes.NewReader(frame.data[4:5]), binary.LittleEndian, &battery_data)
							b.NavData["battery"] = battery_data
							b.sendTelemetry("battery", []byte{battery_data})
						}
					}
				}
			}
		}
	case ARCOMMANDS_ID_PROJECT_ARDRONE3:
		{
			switch commandClass {
			case ARCOMMANDS_ID_ARDRONE3_CLASS_PILOTINGSTATE:
				{
					// This includes things like speed, altitude, GPS coords, and current
					// gross behaviour ("flying"/"landing").
					b.handlePilotingStateFrame(commandId, frame)
				}
			case ARCOMMANDS_ID_ARDRONE3_CLASS_NETWORKSTATE:
				{
					// Wifi scanning!
				}
			}
		}
	default:
		{
			// This shouldn't happen, as there are only two expected projects?
			// Post an unknown telemetry event, may help to discover stuff for future usage.
			go sendTelemetry("unknownProject", frame.data)
		}
	}
}
