package networking

//
// import (
// 	"log"
// 	"net"
// 	"net/http"
// 	_ "net/http/pprof"
// 	"reflect"
// 	"sync"
// 	"syscall"
//
// 	"github.com/gobwas/ws"
// 	"github.com/gobwas/ws/wsutil"
// 	"golang.org/x/sys/unix"
// )
//
// type eventsCollector struct {
// 	fd          int
// 	connections map[int]net.Conn
// 	lock        *sync.RWMutex
// }
//
// var EventsCollector *eventsCollector
//
// func wsHandler(w http.ResponseWriter, r *http.Request) {
// 	// Upgrade connection
// 	conn, _, _, err := ws.UpgradeHTTP(r, w)
// 	if err != nil {
// 		return
// 	}
// 	if err := EventsCollector.Add(conn); err != nil {
// 		log.Printf("Failed to add connection %v", err)
// 		conn.Close()
// 	}
// }
//
// func init() {
// 	// Increase resources limitations
// 	var rLimit syscall.Rlimit
// 	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
// 		panic(err)
// 	}
// 	rLimit.Cur = rLimit.Max
// 	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
// 		panic(err)
// 	}
//
// 	// Enable pprof hooks
// 	go func() {
// 		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
// 			log.Fatalf("pprof failed: %v", err)
// 		}
// 	}()
//
// 	// Start event
// 	var err error
// 	EventsCollector, err = MkEventsCollector()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	go Start()
//
// 	http.HandleFunc("/", wsHandler)
// 	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
// 		log.Fatal(err)
// 	}
// }
//
// func Start() {
// 	for {
// 		connections, err := EventsCollector.Wait()
// 		if err != nil {
// 			log.Printf("Failed to epoll wait %v", err)
// 			continue
// 		}
// 		for _, conn := range connections {
// 			if conn == nil {
// 				break
// 			}
// 			if _, _, err := wsutil.ReadClientData(conn); err != nil {
// 				if err := EventsCollector.Remove(conn); err != nil {
// 					log.Printf("Failed to remove %v", err)
// 				}
// 				conn.Close()
// 			} else {
// 				// This is commented out since in demo usage, stdout is showing messages sent from > 1M connections at very high rate
// 				//log.Printf("msg: %s", string(msg))
// 			}
// 		}
// 	}
// }
// func MkEventsCollector() (*eventsCollector, error) {
// 	fd, err := unix.EpollCreate1(0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &eventsCollector{
// 		fd:          fd,
// 		lock:        &sync.RWMutex{},
// 		connections: make(map[int]net.Conn),
// 	}, nil
// }
//
// func (e *eventsCollector) Add(conn net.Conn) error {
// 	// Extract file descriptor associated with the connection
// 	fd := websocketFD(conn)
// 	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
// 	if err != nil {
// 		return err
// 	}
// 	e.lock.Lock()
// 	defer e.lock.Unlock()
// 	e.connections[fd] = conn
// 	if len(e.connections)%100 == 0 {
// 		log.Printf("Total number of connections: %v", len(e.connections))
// 	}
// 	return nil
// }
//
// func (e *eventsCollector) Remove(conn net.Conn) error {
// 	fd := websocketFD(conn)
// 	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
// 	if err != nil {
// 		return err
// 	}
// 	e.lock.Lock()
// 	defer e.lock.Unlock()
// 	delete(e.connections, fd)
// 	if len(e.connections)%100 == 0 {
// 		log.Printf("Total number of connections: %v", len(e.connections))
// 	}
// 	return nil
// }
//
// func (e *eventsCollector) Wait() ([]net.Conn, error) {
// 	events := make([]unix.EpollEvent, 100)
// 	n, err := unix.EpollWait(e.fd, events, 100)
// 	if err != nil {
// 		return nil, err
// 	}
// 	e.lock.RLock()
// 	defer e.lock.RUnlock()
// 	var connections []net.Conn
// 	for i := 0; i < n; i++ {
// 		conn := e.connections[int(events[i].Fd)]
// 		connections = append(connections, conn)
// 	}
// 	return connections, nil
// }
//
// func websocketFD(conn net.Conn) int {
// 	//tls := reflect.TypeOf(conn.UnderlyingConn()) == reflect.TypeOf(&tls.Conn{})
// 	// Extract the file descriptor associated with the connection
// 	//connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
// 	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
// 	//if tls {
// 	//	tcpConn = reflect.Indirect(tcpConn.Elem())
// 	//}
// 	fdVal := tcpConn.FieldByName("fd")
// 	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
//
// 	return int(pfdVal.FieldByName("Sysfd").Int())
// }
