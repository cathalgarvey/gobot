package bebop

import (
	"github.com/hybridgroup/gobot"
//	"github.com/hybridgroup/gobot/platforms/bebop/client"

//	"github.com/hybridgroup/gobot/platforms/bebop/bbtelem"
)

var _ gobot.Driver = (*BebopDriver)(nil)

// BebopDriver is gobot.Driver representation for the Bebop
type BebopDriver struct {
	name       string
	// connection gobot.Connection
  connection *BebopAdaptor  // Superset of gobot.Connection, so should work?
	gobot.Eventer
	endTelemetry chan struct{}
}

// NewBebopDriver creates an BebopDriver with specified name.
func NewBebopDriver(connection *BebopAdaptor, name string) *BebopDriver {
	d := &BebopDriver{
		name:       name,
		connection: connection,
		Eventer:    gobot.NewEventer(),
		endTelemetry: make(chan struct{}),
	}
	// Gross state telemetry; important enough that this enum got broken out. :)
	d.AddEvent("landed")
	d.AddEvent("takingoff")
	d.AddEvent("hovering")
	d.AddEvent("flying")
	d.AddEvent("landing")
	d.AddEvent("emergency")
	// Introspective telemetry
	d.AddEvent("flattrim")
	d.AddEvent("battery")
	d.AddEvent("autotakeoffmode")
	d.AddEvent("navigatehomestate")
	d.AddEvent("alertstate")
	// Extrospective telemetry
	d.AddEvent("gps")
	d.AddEvent("speed")
	d.AddEvent("attitude")
	d.AddEvent("altitude")
	return d
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
		T := a.connection.Telemetry()
		for {
			select {
			case t := <- T:
				{
					// t is a bbtelem.TelemetryPacket object which may contain error, JSON payload,
					// and/or commentary data in addition to a "Title" property.
					// "Title" is the name of the event to send the JSON payload along.
					gobot.Publish(a.Event(t.Title), t.Payload)
				}
			case <- a.endTelemetry:
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
