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

		fmt.Println("Enabling debug mode.")
		drone.Debug(func(eventname string, payload []byte){
			switch eventname {
				case "flattrim", "camerastate": {
					return
				}
				default: {
					fmt.Println("Event:", eventname, "-", string(payload))
				}
			}

		})

		work := func() {
			fmt.Println("Beginning work.")
      drone.HullProtection(true)
      drone.Land()
			fmt.Println("Pausing before ending work..")
			<- time.After(5*time.Second)
			fmt.Println("Work over.")
    }

		robot := gobot.NewRobot("drone",
        []gobot.Connection{bebopAdaptor},
        []gobot.Device{drone},
        work,
    )

    gbot.AddRobot(robot)
		fmt.Println("Starting.")
    gbot.Start()
}
