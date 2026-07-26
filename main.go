package main

import (
	"flag"
	"log"

	"github.com/devanshu0x/Aster/internal/config"
	"github.com/devanshu0x/Aster/internal/server"
)


func setupFlags(){
	flag.StringVar(&config.HOST,"host","0.0.0.0","host for the aster server")
	flag.IntVar(&config.PORT,"port",6969,"port for the aster server")
	flag.Parse()
}

func main(){
	setupFlags()
	log.Println("Aster is starting...")
	err:=server.RunAsyncTCPServer()
	log.Fatalln("Error closing server: ",err)
}