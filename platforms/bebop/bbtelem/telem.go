package bbtelem

type TelemetryPacket struct{
  Title string
  Comment string
  Error error
  Payload []byte
}
