package server

import "log"

//Run start to run server with db service and router
func Run() {

	r, err := NewRouter()
	if err != nil {
		log.Fatalf("Error on start router: %v", err.Error())
	}
	log.Fatal(r.Run(":8080"))
}
