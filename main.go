package main

import (
	"itodo/app/controllers"
	"itodo/app/models"
	"itodo/config"
)

func main() {
	config.LoadConfig()
	models.CreateDatabase()
	controllers.Requestroute()
}
