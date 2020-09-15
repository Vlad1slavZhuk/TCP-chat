package main

import (
	"log"
	"net"
	"os"
	"tcp-chat/models"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lobby := models.NewLobby()

	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("Listening on localhost:8081.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error:", err)
			continue
		}

		lobby.Join(models.NewUser(conn))
	}
}
