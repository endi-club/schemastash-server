package main

import (
	"fmt"
	"log"

	"github.com/common-nighthawk/go-figure"

	"schemastash/api"
	"schemastash/global"

	"github.com/Valgard/godotenv"
)

func main() {
	fmt.Println(figure.NewFigure("Schemastash", "doom", true).String())
	log.Println("Starting the server...")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	global.ConnectMongo()

	go api.Init()

	log.Println("Server active. Press enter to stop it.")
	fmt.Scanln()
}
