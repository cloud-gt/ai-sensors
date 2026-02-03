package main

import (
	"log"

	"github.com/cloud-gt/ai-sensors/command"
	"github.com/cloud-gt/ai-sensors/manager"
	"github.com/cloud-gt/ai-sensors/server"
)

func main() {
	store := command.NewStore(command.NewMemoryRepository())
	mgr := manager.New(store)
	srv := server.New(store, mgr)

	log.Println("Starting server on :3000")
	if err := srv.ListenAndServe(":3000"); err != nil {
		log.Fatal(err)
	}
}
