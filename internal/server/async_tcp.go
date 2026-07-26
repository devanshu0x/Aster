package server

import (
	"io"
	"log"
	"net"
	"syscall"

	"github.com/devanshu0x/Aster/internal/command"
	"github.com/devanshu0x/Aster/internal/config"
	"github.com/devanshu0x/Aster/internal/resp"
)

type Client struct {
	FD          int
	ReadBuffer  []byte
	WriteBuffer []byte
}

func readSocket(r FDComm, clients map[int]*Client) error {
	buf := make([]byte, 512)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == syscall.EAGAIN {
				break
			}
			log.Println("Failed to read socket")
			return err
		}
		// n == 0, err == nil → the peer has performed an orderly shutdown (closed the connection).
		if n == 0 {
			return io.EOF
		}

		client := clients[r.FD]
		client.ReadBuffer = append(client.ReadBuffer, buf[:n]...)
	}
	return nil
}

func writeSocket(w FDComm, clients map[int]*Client) error {
	client := clients[w.FD]
	if len(client.WriteBuffer) == 0 {
		return nil
	}

	for len(client.WriteBuffer) != 0 {
		writeN, err := w.Write(client.WriteBuffer)
		if err != nil {
			switch err {
			case syscall.EAGAIN:
				return nil

			case syscall.EPIPE, syscall.ECONNRESET:
				return io.EOF

			default:
				return err
			}
		}
		client.WriteBuffer = client.WriteBuffer[writeN:]
	}

	return nil
}

func extractCommand(comm FDComm, clients map[int]*Client) (cmd *resp.RESPValue, done bool, err error) {
	cmd, n, done, err := resp.Decode(clients[comm.FD].ReadBuffer)
	if err != nil {
		return nil, false, err
	}
	if !done {
		return nil, false, nil
	}

	clients[comm.FD].ReadBuffer = clients[comm.FD].ReadBuffer[n:]

	return cmd, true, nil

}

// Asynchronous TCP server (IO Multiplexing)

var con_clients int = 0

func RunAsyncTCPServer() error {
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

	for {
		nevents, err := syscall.EpollWait(epollFD, events, -1)
		if err != nil {
			log.Println("Error: ", err)
			continue
		}

		for i := range nevents {
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

					respVal := command.Dispatch(cmd)

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
