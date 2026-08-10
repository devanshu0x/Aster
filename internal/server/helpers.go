package server

import (
	"io"
	"log"
	"syscall"

	"github.com/devanshu0x/Aster/internal/resp"
)

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