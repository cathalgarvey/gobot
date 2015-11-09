package client

import (
  "bytes"
  "strconv"
  "encoding/binary"
  "encoding/json"
)

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
