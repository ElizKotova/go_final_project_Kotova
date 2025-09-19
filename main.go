package main

import (
	"log"

	"go_final_project_Kotova/pkg/server"
)

func main() {
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
