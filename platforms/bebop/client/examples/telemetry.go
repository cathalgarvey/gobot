package main

import (
    "fmt"
    "time"
    "github.com/hybridgroup/gobot"
    "github.com/hybridgroup/gobot/platforms/bebop"
)

func main() {
    gbot := gobot.NewGobot()
    bebopAdaptor := bebop.NewBebopAdaptor("Drone")
    drone := bebop.NewBebopDriver(bebopAdaptor, "Drone")


		work := func() {
			fmt.Println("Beginning work.")
      drone.HullProtection(true)
  		gobot.On(drone.Event("flying"), func(data interface{}) {
  			gobot.After(1*time.Second, func() {
  				drone.Land()
  			})
  		})
      drone.TakeOff()
    }
    //*
		fmt.Println("Enabling debug mode.")
		drone.Debug(func(eventname string, payload []byte){
			switch eventname {
      case "flattrim", "camerastate":
        {
          // Deliberately avoiding
        }
			  default: {
					fmt.Println("Event:", eventname, "-", string(payload))
				}
			}

		})
    //*/

		robot := gobot.NewRobot("drone",
        []gobot.Connection{bebopAdaptor},
        []gobot.Device{drone},
        work,
    )

    gbot.AddRobot(robot)
		fmt.Println("Starting.")
    gbot.Start()
}
