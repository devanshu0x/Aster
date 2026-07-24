package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/devanshu0x/Aster/internal/config"
)

func readCommand(r io.Reader) (string,error){
	buf:=make([]byte,1024)
	n,err:= r.Read(buf)
	if err!=nil{
		return "",err
	}
	return string(buf[:n]),nil
}

func respond(cmd string, w io.Writer) error{
	 if _,err:=w.Write([]byte(cmd));err!=nil{
		return err
	 }
	 return nil
}

func RunSyncTCPServer(){
	log.Printf("Starting a synchronous TCP server on %s:%d\n",config.HOST,config.PORT)

	conccurrent_clients:=0

	// start listening on configured host:port
	lsnr,err:=net.Listen("tcp",fmt.Sprintf("%s:%d",config.HOST,config.PORT))
	if err!=nil{
		log.Fatalf("Failed to start server: %v",err)
	}

	for{
		// Blocking call: waiting for new client to connect
		c,err:=lsnr.Accept()
		if err!=nil{
			log.Printf("Failed to accept client: %v",err)
			continue
		}

		conccurrent_clients++
		log.Printf("Client connected with address: %s, number of concurrent clients: %d",c.RemoteAddr(),conccurrent_clients)

		for{

			cmd,err:=readCommand(c)
			if err!=nil{
				c.Close()
				conccurrent_clients--
				log.Printf("Client disconnected with address: %s, number of concurrent clients: %d",c.RemoteAddr(),conccurrent_clients)
				if err==io.EOF{
					break
				}
				log.Println("Error: ",err)
			}

			log.Println("Command: ",cmd)
			if err:=respond(cmd,c); err!=nil{
				log.Println("Error Writing: ",err)
			}
		}

	}
}