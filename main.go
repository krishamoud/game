package main

import (
	// Router
	"github.com/krishamoud/game/app/bundles/games"
	"github.com/krishamoud/game/app/common/router"

	// Common code
	_ "github.com/krishamoud/game/app/common/conf"

	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	// close the db connection when we're done
	// defer db.MongoConn.Close()

	// Start loop
	go games.MainGame.GameInterval()
	go games.MainGame.ClientManager.Start(games.MainGame)

	// Handle all requests with gorilla/mux
	http.Handle("/", router.Router())

	// Listen on port 9090
	log.Println("Server listening on port 9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
