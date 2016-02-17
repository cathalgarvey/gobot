package bebop

import (
	"github.com/hybridgroup/gobot"
)

var (
	// TODO: Documentation as to what this achieves would be nice..
	_ gobot.Driver = (*Driver)(nil)
	// BebopEvents is a non-exhaustive list of events that may be posted by
	// an active Bebop object. TODO: Namespace these in-string for sanity?
	BebopEvents = []string{
		// Generic events
		"unknown",
		"unknownProject",
		"error",
		// Gross state telemetry; important enough that this enum got broken out. :)
		"landed",
		"takingoff",
		"hovering",
		"flying",
		"landing",
		"emergency",
		// Introspective telemetry
		/// Camera
		"allstateschanged",
		"camerastate",
		"camerasettingsstate",
		"pictureformatchanged",
		"autowhitebalancechanged",
		"expositionchanged",
		"saturationchanged",
		"timelapsechanged",
		"videoautorecordchanged",
		/// Behaviour
		"maxaltitudechanged",
		"maxtiltchanged",
		"absolutcontrolchanged",
		"maxdistancechanged",
		"noflyovermaxdistancechanged",
		"maxverticalspeedchanged",
		"maxrotationspeedchanged",
		"hullprotectionchanged",
		"outdoorchanged",
		"flattrim",
		"navigatehomestate",
		"alertstate",
		"autotakeoffmode",
		"networksettingsstate",
		"mavlinkfileplaying",
		"availabilitystatechanged",
		"startingerrorevent",
		"speedbridleevent",
		"sethomechanged",
		"resethomechanged",
		"gpsfixstatechanged",
		"gpsupdatestatechanged",
		"hometypechanged",
		/// Network
		"networkdisconnect",
		"wifiscanlistchanged",        // Sent one for each wifi scanned?
		"allwifiscanchanged",         // Sent when above frames are all sent
		"wifiauthchannellistchanged", // Sent to indicate for 2.4/5ghz channels which are permitted, and where (indoors/outdoors)
		"allwifiauthchannelchanged",  // Sent to indicate device is finished sending the above?
		/// Assets
		"battery",
		"massstorage",
		"massstorageinfo",
		"massstorageinforemaining",
		"sensorstates",
		/// Factoids
		"currentdate",
		"currenttime",
		"dronemodel",
		"countrycodes",
		"controllerlibversion",
		"skycontrollerlibversion",
		"devicelibversion",
		// Extrospective telemetry
		"gps",
		"speed",
		"attitude",
		"altitude",
		"wifisignal",
	}
)

// Driver is gobot.Driver representation for the Bebop
type Driver struct {
	name       string
	connection gobot.Connection
	gobot.Eventer
	endTelemetry chan struct{}
}

// NewBebopDriver creates an BebopDriver with specified name.
func NewBebopDriver(connection *BebopAdaptor, name string) *Driver {
	d := &Driver{
		name:         name,
		connection:   connection,
		Eventer:      gobot.NewEventer(),
		endTelemetry: make(chan struct{}),
	}
	for _, e := range BebopEvents {
		e := e
		d.AddEvent(e)
	}
	return d
}

// Debug registers a given function to subscribe to all known Bebop Events,
// including "unknown" and "error"
func (a *Driver) Debug(f func(string, []byte)) {
	for _, e := range BebopEvents {
		e := e
		gobot.On(a.Event(e), func(data interface{}) {
			switch t := data.(type) {
			case []byte:
				{
					f(e, t)
				}
			default:
				{
					// Avoid killing telemetry by mistake when sending events manually
					f(e, []byte(`{"warning":"Manual telemetry event may break things"}`))
				}
			}
		})
	}
}

// Name returns the Drivers Name
func (a *Driver) Name() string { return a.name }

// Connection returns the Drivers Connection
func (a *Driver) Connection() gobot.Connection { return a.connection }

// adaptor returns ardrone adaptor
func (a *Driver) adaptor() *BebopAdaptor {
	return a.Connection().(*BebopAdaptor)
}

// StopTelemetry stops further telemetry messages being posted
// to the event queue, and closes the channel from the adaptor also.
func (a *Driver) StopTelemetry() error {
	close(a.endTelemetry)
	return a.adaptor().drone.StopTelemetry()
}

// Start starts the Driver.
// This spins out a goroutine to read telemetry from the adaptor and post
// events;
func (a *Driver) Start() (errs []error) {
	go func(a *Driver) {
		T := a.adaptor().drone.Telemetry()
		for {
			select {
			case <-a.endTelemetry:
				{
					return
				}
			case t := <-T:
				{
					// t is a bbtelem.TelemetryPacket object which may contain error, JSON payload,
					// and/or commentary data in addition to a "Title" property.
					// "Title" is the name of the event to send the JSON payload along.
					if t.Title == "error" || t.Title == "unknown" {
						var payload []byte
						payload = append(payload, []byte(t.Comment)...)
						payload = append(payload, []byte(":: ")...)
						if t.Error != nil {
							payload = append(payload, []byte(t.Error.Error())...)
						}
						payload = append(payload, t.Payload...)
						gobot.Publish(a.Event(t.Title), payload)
					} else {
						gobot.Publish(a.Event(t.Title), t.Payload)
					}
				}
			}
		}
	}(a)
	return
}

// Halt halts the Driver
func (a *Driver) Halt() (errs []error) {
	// TODO: ?
	return
}

// TakeOff makes the drone start flying
func (a *Driver) TakeOff() {
	a.adaptor().drone.TakeOff()
	// "flying" event should be published by usual event handling system, now
	// however, it only publishes *after* "takingoff"!
	//gobot.Publish(a.Event("flying"), a.adaptor().drone.TakeOff())
}

// Land causes the drone to land
func (a *Driver) Land() {
	// TODO: Why is this broken?
	a.adaptor().drone.Land()
}

// Up makes the drone gain altitude.
// speed can be a value from `0` to `100`.
func (a *Driver) Up(speed int) {
	a.adaptor().drone.Up(speed)
}

// Down makes the drone reduce altitude.
// speed can be a value from `0` to `100`.
func (a *Driver) Down(speed int) {
	a.adaptor().drone.Down(speed)
}

// Left causes the drone to bank to the left, controls the roll, which is
// a horizontal movement using the camera as a reference point.
// speed can be a value from `0` to `100`.
func (a *Driver) Left(speed int) {
	a.adaptor().drone.Left(speed)
}

// Right causes the drone to bank to the right, controls the roll, which is
// a horizontal movement using the camera as a reference point.
// speed can be a value from `0` to `100`.
func (a *Driver) Right(speed int) {
	a.adaptor().drone.Right(speed)
}

// Forward causes the drone go forward, controls the pitch.
// speed can be a value from `0` to `100`.
func (a *Driver) Forward(speed int) {
	a.adaptor().drone.Forward(speed)
}

// Backward causes the drone go forward, controls the pitch.
// speed can be a value from `0` to `100`.
func (a *Driver) Backward(speed int) {
	a.adaptor().drone.Backward(speed)
}

// Clockwise causes the drone to spin in clockwise direction
// speed can be a value from `0` to `100`.
func (a *Driver) Clockwise(speed int) {
	a.adaptor().drone.Clockwise(speed)
}

// CounterClockwise the drone to spin in counter clockwise direction
// speed can be a value from `0` to `100`.
func (a *Driver) CounterClockwise(speed int) {
	a.adaptor().drone.CounterClockwise(speed)
}

// Stop makes the drone to hover in place.
func (a *Driver) Stop() {
	a.adaptor().drone.Stop()
	close(a.endTelemetry)
}

// Video returns a channel which raw video frames will be broadcast on
func (a *Driver) Video() chan []byte {
	return a.adaptor().drone.Video()
}

// StartRecording starts the recording video to the drones interal storage
func (a *Driver) StartRecording() error {
	return a.adaptor().drone.StartRecording()
}

// StopRecording stops a previously started recording
func (a *Driver) StopRecording() error {
	return a.adaptor().drone.StopRecording()
}

// HullProtection tells the drone if the hull/prop protectors are attached. This is needed to adjust flight characteristics of the Bebop.
func (a *Driver) HullProtection(protect bool) error {
	return a.adaptor().drone.HullProtection(protect)
}

// Outdoor tells the drone if flying Outdoor or not. This is needed to adjust flight characteristics of the Bebop.
func (a *Driver) Outdoor(outdoor bool) error {
	return a.adaptor().drone.Outdoor(outdoor)
}
