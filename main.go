package main

import (
	"log"

	"github.com/cloud-gt/ai-sensors/command"
	"github.com/cloud-gt/ai-sensors/dashboard"
	"github.com/cloud-gt/ai-sensors/manager"
	"github.com/cloud-gt/ai-sensors/server"
)

func main() {
	store := command.NewStore(command.NewMemoryRepository())
	mgr := manager.New(store)
	srv := server.New(store, mgr)

	dashFS, err := dashboard.FS()
	if err != nil {
		log.Fatal("failed to load dashboard assets: ", err)
	}
	srv.MountDashboard(dashFS)

	log.Println("Starting server on :3000")
	log.Println("Dashboard available at http://localhost:3000/dashboard")
	if err := srv.ListenAndServe(":3000"); err != nil {
		log.Fatal(err)
	}
}
