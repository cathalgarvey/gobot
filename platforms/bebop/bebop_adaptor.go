package bebop

import (
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/bebop/client"
	"github.com/hybridgroup/gobot/platforms/bebop/bbtelem"
)

var _ gobot.Adaptor = (*BebopAdaptor)(nil)

// drone defines expected drone behaviour
type drone interface {
	TakeOff() error
	Land() error
	Up(n int) error
	Down(n int) error
	Left(n int) error
	Right(n int) error
	Forward(n int) error
	Backward(n int) error
	Clockwise(n int) error
	CounterClockwise(n int) error
	Stop() error
	Connect() error
	Video() chan []byte
	// Returns a channel of events as returned by the device, with string titles.
	// Titles are friendly in some cases, and SDK allcaps identifiers in others.
	// Returns a channel of JSON events as returned by the device, each of which
	// will have a field "TITLE" (corresponding to a string) as well as other data
	// fields according to event type. The TITLE field is taken as the event Name
	// and dispatched to the gobot event handler.
	Telemetry() chan bbtelem.TelemetryPacket
	StartRecording() error
	StopRecording() error
	HullProtection(protect bool) error
	Outdoor(outdoor bool) error
}

// BebopAdaptor is gobot.Adaptor representation for the Bebop
type BebopAdaptor struct {
	name  string
	drone drone
	//config  client.Config
	connect func(*BebopAdaptor) error
}

// NewBebopAdaptor returns a new BebopAdaptor
func NewBebopAdaptor(name string) *BebopAdaptor {
	return &BebopAdaptor{
		name:  name,
		drone: client.New(),
		connect: func(a *BebopAdaptor) error {
			return a.drone.Connect()
		},
	}
}

// Name returns the BebopAdaptors Name
func (a *BebopAdaptor) Name() string { return a.name }

// Connect establishes a connection to the ardrone
func (a *BebopAdaptor) Connect() (errs []error) {
	err := a.connect(a)
	if err != nil {
		return []error{err}
	}
	return
}

func (a *BebopAdaptor) Telemetry() chan bbtelem.TelemetryPacket {
	return a.drone.Telemetry()
}

// Finalize terminates the connection to the ardrone
func (a *BebopAdaptor) Finalize() (errs []error) { return }
