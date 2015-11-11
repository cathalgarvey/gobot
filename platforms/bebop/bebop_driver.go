package bebop

import (
	//"fmt"
	"github.com/hybridgroup/gobot"
)

var (
	_           gobot.Driver = (*BebopDriver)(nil)
	bebopEvents              = []string{
		// Generic events
		"unknown",
		"unknownProject", // So common it merits its own.. handling code must be picking up non data frames?
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
		"allstateschanged"
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
		"wifiscanlistchanged", // Sent one for each wifi scanned?
		"allwifiscanchanged",  // Sent when above frames are all sent
		"wifiauthchannellistchanged",  // Sent to indicate for 2.4/5ghz channels which are permitted, and where (indoors/outdoors)
		"allwifiauthchannelchanged", // Sent to indicate device is finished sending the above?
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

// BebopDriver is gobot.Driver representation for the Bebop
type BebopDriver struct {
	name       string
	connection gobot.Connection
	gobot.Eventer
	endTelemetry chan struct{}
}

// NewBebopDriver creates an BebopDriver with specified name.
func NewBebopDriver(connection *BebopAdaptor, name string) *BebopDriver {
	d := &BebopDriver{
		name:         name,
		connection:   connection,
		Eventer:      gobot.NewEventer(),
		endTelemetry: make(chan struct{}),
	}
	for _, e := range bebopEvents {
		e := e
		d.AddEvent(e)
	}
	return d
}

// Use given function to subscribe to all known Bebop Events,
// including "unknown" and "error"
func (a *BebopDriver) Debug(f func(string, []byte)) {
	for _, e := range bebopEvents {
		e := e
		gobot.On(a.Event(e), func(data interface{}) {
			switch t := data.(type) {
				case []byte: {
						f(e, t)
					}
				default: {
				  // Avoid killing telemetry by mistake when sending events manually
					f(e, []byte(`{"warning":"Manual telemetry event may break things"}`))
				}
			}
		})
	}
}

// Name returns the BebopDrivers Name
func (a *BebopDriver) Name() string { return a.name }

// Connection returns the BebopDrivers Connection
func (a *BebopDriver) Connection() gobot.Connection { return a.connection }

// adaptor returns ardrone adaptor
func (a *BebopDriver) adaptor() *BebopAdaptor {
	return a.Connection().(*BebopAdaptor)
}

// Start starts the BebopDriver.
// This spins out a goroutine to read telemetry from the adaptor and post
// events;
func (a *BebopDriver) Start() (errs []error) {
	go func(a *BebopDriver) {
		T := a.adaptor().drone.Telemetry()
		for {
			select {
			case t := <-T:
				{
					// t is a bbtelem.TelemetryPacket object which may contain error, JSON payload,
					// and/or commentary data in addition to a "Title" property.
					// "Title" is the name of the event to send the JSON payload along.
					if t.Title == "error" || t.Title == "unknown" {
						payload := make([]byte, 0)
						payload = append(payload, []byte(t.Comment)...)
						payload = append(payload, []byte(":: ")...)
						if t.Error != nil {
							payload = append(payload, []byte(t.Error.Error())...)
						}
						payload = append(payload, t.Payload...)
						gobot.Publish(a.Event(t.Title), payload)
					} else {
						// fmt.Println("Issuing telemetry: ", t.Title)
						gobot.Publish(a.Event(t.Title), t.Payload)
					}
				}
			case <-a.endTelemetry:
				{
					return
				}
			}
		}
	}(a)
	return
}

// Halt halts the BebopDriver
func (a *BebopDriver) Halt() (errs []error) {
	return
}

// TakeOff makes the drone start flying
func (a *BebopDriver) TakeOff() {
	a.adaptor().drone.TakeOff()
	// "flying" event should be published by usual event handling system, now?
	// But..it isn't?
	gobot.Publish(a.Event("flying"), a.adaptor().drone.TakeOff())
}

// Land causes the drone to land
func (a *BebopDriver) Land() {
	a.adaptor().drone.Land()
}

// Up makes the drone gain altitude.
// speed can be a value from `0` to `100`.
func (a *BebopDriver) Up(speed int) {
	a.adaptor().drone.Up(speed)
}

// Down makes the drone reduce altitude.
// speed can be a value from `0` to `100`.
func (a *BebopDriver) Down(speed int) {
	a.adaptor().drone.Down(speed)
}

// Left causes the drone to bank to the left, controls the roll, which is
// a horizontal movement using the camera as a reference point.
// speed can be a value from `0` to `100`.
func (a *BebopDriver) Left(speed int) {
	a.adaptor().drone.Left(speed)
}

// Right causes the drone to bank to the right, controls the roll, which is
// a horizontal movement using the camera as a reference point.
// speed can be a value from `0` to `100`.
func (a *BebopDriver) Right(speed int) {
	a.adaptor().drone.Right(speed)
}

// Forward causes the drone go forward, controls the pitch.
// speed can be a value from `0` to `100`.
func (a *BebopDriver) Forward(speed int) {
	a.adaptor().drone.Forward(speed)
}

// Backward causes the drone go forward, controls the pitch.
// speed can be a value from `0` to `100`.
func (a *BebopDriver) Backward(speed int) {
	a.adaptor().drone.Backward(speed)
}

// Clockwise causes the drone to spin in clockwise direction
// speed can be a value from `0` to `100`.
func (a *BebopDriver) Clockwise(speed int) {
	a.adaptor().drone.Clockwise(speed)
}

// CounterClockwise the drone to spin in counter clockwise direction
// speed can be a value from `0` to `100`.
func (a *BebopDriver) CounterClockwise(speed int) {
	a.adaptor().drone.CounterClockwise(speed)
}

// Stop makes the drone to hover in place.
func (a *BebopDriver) Stop() {
	a.adaptor().drone.Stop()
	close(a.endTelemetry)
}

// Video returns a channel which raw video frames will be broadcast on
func (a *BebopDriver) Video() chan []byte {
	return a.adaptor().drone.Video()
}

// StartRecording starts the recording video to the drones interal storage
func (a *BebopDriver) StartRecording() error {
	return a.adaptor().drone.StartRecording()
}

// StopRecording stops a previously started recording
func (a *BebopDriver) StopRecording() error {
	return a.adaptor().drone.StopRecording()
}

// HullProtection tells the drone if the hull/prop protectors are attached. This is needed to adjust flight characteristics of the Bebop.
func (a *BebopDriver) HullProtection(protect bool) error {
	return a.adaptor().drone.HullProtection(protect)
}

// Outdoor tells the drone if flying Outdoor or not. This is needed to adjust flight characteristics of the Bebop.
func (a *BebopDriver) Outdoor(outdoor bool) error {
	return a.adaptor().drone.Outdoor(outdoor)
}
