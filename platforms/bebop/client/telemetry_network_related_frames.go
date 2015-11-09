package client

import (
  "bytes"
  "strconv"
  "encoding/binary"
  "encoding/json"
)
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
