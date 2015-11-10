package client

import (
  "bytes"
  "strconv"
  "encoding/binary"
  "encoding/json"
)


// Internal states, like settings, battery level, storage, date/time,
func (b *Bebop) handleCommonStateFrame(commandId byte, frame *NetworkFrame) {
	switch commandId {
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_ALLSTATESCHANGED:
		{
			// Is this even useful telemetry?
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
      b.sendJSONTelemetry(frame, "massstorage", struct {
				Mass_storage_id uint8  `json:"Mass_storage_id"`
				Name            string `json:"name"`
			}{Mass_storage_id: mass_storage_id, Name: mass_storage_name})
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
      b.sendJSONTelemetry(frame, "massstorageinfo", telemdata)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTDATECHANGED:
		// Date in ISO-8601
		{
			dates := string(frame.Data[4:]) // Parse to real time object? ISO-8601
      b.sendJSONTelemetry(frame, "currentdate", struct {
				Date string `json:"date"`
			}{Date: dates})
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_CURRENTTIMECHANGED:
		// Time in ISO-8601
		{
			times := string(frame.Data[4:]) // Parse to real time object? ISO-8601
      b.sendJSONTelemetry(frame, "currenttime", struct {
				Time string `json:"time"`
			}{Time: times})
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
      b.sendJSONTelemetry(frame, "massstorageinforemaining", telemdata)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_WIFISIGNALCHANGED:
		{
			var telemdata struct {
				Rssi int16 `json:"rssi"`
			} // in dbm
			binary.Read(bytes.NewReader(frame.Data[4:20]), binary.LittleEndian, &telemdata)
      b.sendJSONTelemetry(frame, "wifisignal", telemdata)
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_SENSORSSTATESLISTCHANGED:
		{
			var sensorState bool
			sensorName, err := decodeEnum(frame.Data[4:8], []string{"IMU", "barometer", "ultrasound", "GPS", "magnetometer", "vertical_camera"})
			if err != nil {
				go b.sendRuntimeError("Error processing sensor state telemetry", err, frame.Data)
				return
			}
      b.sendJSONTelemetry(frame, "sensorstates", struct{
				SensorName  string `json:"sensorName"`
				SensorState bool   `json:"sensorState"`
			}{SensorName: sensorName, SensorState: sensorState})
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_PRODUCTMODEL:
		// This appears to be irrelevant to the Bebop but it's in "common"!
		{
			modelstr, err := decodeEnum(frame.Data[4:8], []string{"RS_TRAVIS", "RS_MARS", "RS_SWAT", "RS_MCLANE", "RS_BLAZE", "RS_ORAK", "RS_NEWZ", "JS_DIESEL", "JS_BUZZ", "JS_MAX", "JS_JETT", "JS_TUKTUK"})
			if err != nil {
				go b.sendRuntimeError("Error processing drone model telemetry", err, frame.Data)
				return
			}
      b.sendJSONTelemetry(frame, "dronemodel", struct {
				Model string `json:"model"`
			}{Model: modelstr})
		}
	case ARCOMMANDS_ID_COMMON_COMMONSTATE_CMD_COUNTRYLISTKNOWN:
		{
			ccodes := string(frame.Data[4:])
      b.sendJSONTelemetry(frame, "countrycodes", struct {
				CountryCodes string `json:"countryCodes"`
			}{ccodes})
		}
	default:
		{
			go b.sendUnknownTelemetry("Unknown/Unhandled COMMONSTATE commandId: "+strconv.Itoa(int(commandId)), frame.Data)
		}
	}
}
