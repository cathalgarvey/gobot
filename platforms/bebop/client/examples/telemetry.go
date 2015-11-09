package main

import (
	"fmt"
	"time"
	"github.com/hybridgroup/gobot/platforms/bebop/client"
)

func main() {
	bebop := client.New()

	if err := bebop.Connect(); err != nil {
		fmt.Println(err)
		return
	}

  bebop.Debug(func(eventname string, payload []byte){
		fmt.Println("Event:", eventname, "-", string(payload))
	})

	fmt.Println("hull")
	bebop.HullProtection(true)
}
