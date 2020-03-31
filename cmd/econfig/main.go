package main

import (
	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/device"
	uuid "github.com/satori/go.uuid"
	"log"
)

func main() {
	cloudCxt := &controller.CloudCtx{}
	devID, _ := uuid.NewV4()
	err := cloudCxt.AddDevice(&devID, device.DevModelQemu)
	if err != nil {
		log.Fatal(err)
	}
	b, err := cloudCxt.GetConfigBytes(&devID)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(string(b))
}
