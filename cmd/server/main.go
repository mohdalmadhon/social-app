package main

import (
	"log"
	"net/http"
	server "social/cmd"
	"social/cmd/server/routes"
	"social/cmd/utils"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := server.ConnectToDB("sqlite3", "./db/social_network.db?_foreign_keys=on")
	if err != nil {
		log.Println(err)
		return
	}

	if err := utils.MakeUploadDirectories(); err != nil {
		log.Println(err)
		return
	}

	if err := utils.SeedUploads(); err != nil {
		log.Println(err)
		return
	}

	if err := utils.CopyDefaultAvatar(); err != nil {
		log.Println(err)
		return
	}
	
	server := http.Server{
		Handler: routes.StartServer(db),
		Addr:    ":4000",
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Println("ERROR: STARTING SERVER", err)
		return
	}
	log.Println("SERVER STARTED ON POST: " + server.Addr)
}