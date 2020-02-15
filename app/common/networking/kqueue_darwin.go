// +build darwin

package networking

import (
	"log"
	"reflect"
	"sync"
	"syscall"

	"github.com/gobwas/ws/wsutil"

	"net"
)

type EventsCollector struct {
	fd          int
	connections map[int]net.Conn
	// kqueue will watch these Kevent_t changes after call Kevent()
	// see more in freeBSD paper: https://people.freebsd.org/~jlemon/papers/kqueue.pdf
	changes []syscall.Kevent_t
	lock    *sync.RWMutex
}

func MkEventsCollector() (*EventsCollector, error) {
	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}
	kevent := syscall.Kevent_t{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}
	if _, err = syscall.Kevent(fd, []syscall.Kevent_t{kevent}, nil, nil); err != nil {
		return nil, err
	}
	return &EventsCollector{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

func (e *EventsCollector) Add(conn net.Conn) (error, int) {
	fd := websocketFD(conn)
	e.changes = append(e.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
	)
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	return nil, fd
}

func (e *EventsCollector) GetFdFromConnection(conn net.Conn) int {
	return websocketFD(conn)
}

func (e *EventsCollector) Remove(conn net.Conn) error {
	fd := websocketFD(conn)
	e.changes = append(e.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_READ,
		},
	)
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	return nil
}

func (e *EventsCollector) Wait() ([]net.Conn, error) {
	events := make([]syscall.Kevent_t, 100)
	n, err := syscall.Kevent(e.fd, e.changes, events, nil)
	if err != nil {
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Ident)]
		connections = append(connections, conn)
	}
	return connections, nil
}
func (e *EventsCollector) Start() {
	for {
		connections, err := e.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if _, _, err := wsutil.ReadClientData(conn); err != nil {
				if err := e.Remove(conn); err != nil {
					log.Printf("Failed to remove %v", err)
				}
				conn.Close()
			} else {
				// This is commented out since in demo usage, stdout is showing messages sent from > 1M connections at very high rate
				//log.Printf("msg: %s", string(msg))
			}
		}
	}
}
func websocketFD(conn net.Conn) int {
	//tls := reflect.TypeOf(conn.UnderlyingConn()) == reflect.TypeOf(&tls.Conn{})
	// Extract the file descriptor associated with the connection
	//connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	//if tls {
	//	tcpConn = reflect.Indirect(tcpConn.Elem())
	//}
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}
