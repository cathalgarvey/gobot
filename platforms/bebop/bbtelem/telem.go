package bbtelem

type TelemetryPacket struct {
	Title   string
	Comment string
	Error   error
	Payload []byte
}

// TODO: Convert client and driver to use the following variables instead of
// raw strings.

var (
	Unknown        = "bebop:unknown"
	Unknownproject = "bebop:unknownProject"
	Error          = "bebop:error"

	// Gross state telemetry; important enough that this enum got broken out. :)
	Landed    = "bebop:landed"
	Takingoff = "bebop:takingoff"
	Hovering  = "bebop:hovering"
	Flying    = "bebop:flying"
	Landing   = "bebop:landing"
	Emergency = "bebop:emergency"

	// Introspective telemetry

	/// Camera
	Allstateschanged        = "bebop:allstateschanged"
	Camerastate             = "bebop:camerastate"
	Camerasettingsstate     = "bebop:camerasettingsstate"
	Pictureformatchanged    = "bebop:pictureformatchanged"
	Autowhitebalancechanged = "bebop:autowhitebalancechanged"
	Expositionchanged       = "bebop:expositionchanged"
	Saturationchanged       = "bebop:saturationchanged"
	Timelapsechanged        = "bebop:timelapsechanged"
	Videoautorecordchanged  = "bebop:videoautorecordchanged"

	/// Behaviour
	Maxaltitudechanged          = "bebop:maxaltitudechanged"
	Maxtiltchanged              = "bebop:maxtiltchanged"
	Absolutcontrolchanged       = "bebop:absolutcontrolchanged"
	Maxdistancechanged          = "bebop:maxdistancechanged"
	Noflyovermaxdistancechanged = "bebop:noflyovermaxdistancechanged"
	Maxverticalspeedchanged     = "bebop:maxverticalspeedchanged"
	Maxrotationspeedchanged     = "bebop:maxrotationspeedchanged"
	Hullprotectionchanged       = "bebop:hullprotectionchanged"
	Outdoorchanged              = "bebop:outdoorchanged"
	Flattrim                    = "bebop:flattrim"
	Navigatehomestate           = "bebop:navigatehomestate"
	Alertstate                  = "bebop:alertstate"
	Autotakeoffmode             = "bebop:autotakeoffmode"
	Networksettingsstate        = "bebop:networksettingsstate"
	Mavlinkfileplaying          = "bebop:mavlinkfileplaying"
	Availabilitystatechanged    = "bebop:availabilitystatechanged"
	Startingerrorevent          = "bebop:startingerrorevent"
	Speedbridleevent            = "bebop:speedbridleevent"
	Sethomechanged              = "bebop:sethomechanged"
	Resethomechanged            = "bebop:resethomechanged"
	Gpsfixstatechanged          = "bebop:gpsfixstatechanged"
	Gpsupdatestatechanged       = "bebop:gpsupdatestatechanged"
	Hometypechanged             = "bebop:hometypechanged"

	/// Network
	Networkdisconnect          = "bebop:networkdisconnect"
	Wifiscanlistchanged        = "bebop:wifiscanlistchanged"
	Allwifiscanchanged         = "bebop:allwifiscanchanged"
	Wifiauthchannellistchanged = "bebop:wifiauthchannellistchanged"
	Allwifiauthchannelchanged  = "bebop:allwifiauthchannelchanged"

	/// Assets
	Battery                  = "bebop:battery"
	Massstorage              = "bebop:massstorage"
	Massstorageinfo          = "bebop:massstorageinfo"
	Massstorageinforemaining = "bebop:massstorageinforemaining"
	Sensorstates             = "bebop:sensorstates"

	/// Factoids
	Currentdate             = "bebop:currentdate"
	Currenttime             = "bebop:currenttime"
	Dronemodel              = "bebop:dronemodel"
	Countrycodes            = "bebop:countrycodes"
	Controllerlibversion    = "bebop:controllerlibversion"
	Skycontrollerlibversion = "bebop:skycontrollerlibversion"
	Devicelibversion        = "bebop:devicelibversion"

	// Extrospective telemetry
	Gps        = "bebop:gps"
	Speed      = "bebop:speed"
	Attitude   = "bebop:attitude"
	Altitude   = "bebop:altitude"
	Wifisignal = "bebop:wifisignal"
)
