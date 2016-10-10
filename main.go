package main

import (
	"github.com/reechou/x-real-control/config"
	"github.com/reechou/x-real-control/controller"
)

func main() {
	controller.NewControllerLogic(config.NewConfig()).Start()
}
