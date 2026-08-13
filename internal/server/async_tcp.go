package server

import (
	"io"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/devanshu0x/Aster/internal/command"
	"github.com/devanshu0x/Aster/internal/config"
	"github.com/devanshu0x/Aster/internal/persistence"
	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

// Asynchronous TCP server (IO Multiplexing)

var con_clients int = 0
var cronFrequency time.Duration= 1*time.Second
var lastCronExecTime time.Time= time.Now()

func RunAsyncTCPServer(aof *persistence.AOF) error {
	log.Printf("Starting TCP server on %s:%d\n", config.HOST, config.PORT)

	clients := map[int]*Client{}

	// This is not the maximum number of clients server can actually support
	// This is simply the maximum number of events epoll_wait() can return in a single call
	max_clients := 200000

	// epoll event slice to hold events
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	//creating a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	defer syscall.Close(serverFD)

	// Set the socket to operate in non-blocking mode
	// if err= syscall.SetNonblock(serverFD,true); err!=nil{
	// 	return err
	// }

	// bind the ip and port
	ip4 := net.ParseIP(config.HOST)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.PORT,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	// start listening
	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	// Async IO or IO Multiplexing starts here

	// create an epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}

	defer syscall.Close(epollFD)

	// Here we will specify the events we want to be notified for
	var serverSocketEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// listen to read events on the server itself
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &serverSocketEvent); err != nil {
		return err
	}
	// event loop
	for {
		timeout:=int(time.Until(lastCronExecTime.Add(cronFrequency)).Milliseconds())
		if timeout<0{
			timeout=0
		}

		nevents, err := syscall.EpollWait(epollFD, events, timeout)
		if err != nil {
			log.Println("Error: ", err)
			continue
		}

		// active deletion of expired keys
		if time.Now().After(lastCronExecTime.Add(cronFrequency)){
			store.DeleteExpiredKeys()
			lastCronExecTime= time.Now()
		}

		for i:=0;i<nevents;i++ {
			// if the server socket itself is ready for an IO
			if events[i].Fd == int32(serverFD) {
				// accept incoming connection from a client
				// read all incoming connections
				for {
					fd, _, err := syscall.Accept(serverFD)
					if err != nil {
						if err == syscall.EAGAIN {
							break
						}
						log.Println("Error: ", err)
						break
					}

					clients[fd] = &Client{
						FD: fd,
					}
					con_clients++
					syscall.SetNonblock(fd, true)

					// add this new tcp client to be monitored
					var clientSocketEvent syscall.EpollEvent = syscall.EpollEvent{
						Events: syscall.EPOLLIN,
						Fd:     int32(fd),
					}
					if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &clientSocketEvent); err != nil {
						return err
					}
					log.Printf("Client connected (fd=%d, %d clients connected)", fd, con_clients)
				}

			} else {
				comm := FDComm{FD: int(events[i].Fd)}
				// cmd,err:=readCommand(comm)
				// if err!=nil{
				// 	syscall.Close(int(events[i].Fd))
				// 	con_clients--
				// 	continue
				// }

				// respond(cmd,comm)
				if err := readSocket(comm, clients); err != nil {
					if err == io.EOF {
						// closing the fd gracefully
						if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_DEL, comm.FD, nil); err != nil {
							return err
						}
						if err := syscall.Close(comm.FD); err != nil {
							return err
						}
						delete(clients, comm.FD)
						con_clients--
						log.Printf("Client disconnected (%d clients connected)", con_clients)
						continue
					}
					return err
				}

				for {
					cmd, done, err := extractCommand(comm, clients)
					if err != nil {
						log.Println("Error while extracting command: ", err)
						break
					}
					if !done {
						break
					}

					respVal,is_mutated := command.Dispatch(cmd)
					if is_mutated && aof!=nil{
						if err:=aof.Append(cmd);err!=nil{
							return err
						}
					}

					encoding, err := resp.Encode(respVal)
					if err != nil {
						return err
					}
					clients[comm.FD].WriteBuffer = append(clients[comm.FD].WriteBuffer, encoding...)

					if err := writeSocket(comm, clients); err != nil {
						if err == io.EOF {
							// closing the fd gracefully
							if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_DEL, comm.FD, nil); err != nil {
								return err
							}
							if err := syscall.Close(comm.FD); err != nil {
								return err
							}
							delete(clients, comm.FD)
							con_clients--
							log.Printf("Client disconnected (%d clients connected)", con_clients)
							break
						}
						return err
					}

				}

			}
		}
	}

}
