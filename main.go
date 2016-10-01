package main

import (
	"github.com/reechou/x-real-control/controller"
)

func main() {
	controller.NewControllerLogic(controller.NewConfig()).Start()
}
