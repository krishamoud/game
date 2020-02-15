package main

import (
	// Router

	"syscall"

	"github.com/krishamoud/game/app/bundles/games"
	"github.com/krishamoud/game/app/common/router"

	// Common code
	_ "github.com/krishamoud/game/app/common/conf"

	"net/http"
	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"
)

func main() {
	// close the db connection when we're done
	// defer db.MongoConn.Close()

	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	// Enable pprof hooks
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	// Start loop
	games.MainGame.SetupEventsCollector()
	go games.MainGame.GameInterval()
	go games.MainGame.Hub.Start()
	// go games.MainGame.ClientManager.Start(games.MainGame)
	// Start event
	// var err error
	// EventsCollector, err := networking.MkEventsCollector()
	// if err != nil {
	// 	panic(err)
	// }

	// go EventsCollector.Start()

	// Handle all requests with gorilla/mux
	http.Handle("/", router.Router())

	// Listen on port 9090
	log.Println("Server listening on port 9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
