package main

import (
	"github.com/reechou/x-real-control/controller"
	"github.com/reechou/x-real-control/config"
)

func main() {
	controller.NewControllerLogic(config.NewConfig()).Start()
}
