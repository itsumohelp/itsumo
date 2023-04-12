package main

import (
	"fmt"
	"itodo/app/controllers"
	"itodo/app/models"
	"itodo/config"
)

func main() {
	fmt.Println("a")
	config.LoadConfig()
	fmt.Println("b")
	models.InitDataBase()
	fmt.Println("c")
	controllers.Requestroute()
}
