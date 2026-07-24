package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/devanshu0x/Aster/internal/config"
)

func readCommand(r io.Reader) (string,error){
	/*
	Here Rigt now I'm assuming that the entire command will be read in a single
	read call, but we know its not guaranteed by tcp, we will work on that later
	*/
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

	concurrent_clients:=0

	// start listening on configured host:port
	lsnr,err:=net.Listen("tcp",fmt.Sprintf("%s:%d",config.HOST,config.PORT))
	if err!=nil{
		log.Fatalf("Failed to start server: %v",err)
	}

	defer lsnr.Close()

	for{
		// Blocking call: waiting for new client to connect
		c,err:=lsnr.Accept()
		if err!=nil{
			log.Printf("Failed to accept client: %v",err)
			continue
		}

		concurrent_clients++
		log.Printf("Client connected with address: %s, number of concurrent clients: %d",c.RemoteAddr(),concurrent_clients)

		for{

			cmd,err:=readCommand(c)
			if err!=nil{
				c.Close()
				concurrent_clients--
				log.Printf("Client disconnected with address: %s, number of concurrent clients: %d",c.RemoteAddr(),concurrent_clients)
				if err==io.EOF{
					break
				}
				log.Println("Error: ",err)
				break
			}

			log.Println("Command: ",cmd)
			if err:=respond(cmd,c); err!=nil{
				log.Println("Error Writing: ",err)
				break
			}
		}

	}
}