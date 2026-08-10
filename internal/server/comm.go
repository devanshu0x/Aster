package server

import "syscall"

// File Descriptor Communicator
type FDComm struct{
	FD int
}

func (f FDComm) Write(b []byte) (int, error){
	return syscall.Write(f.FD,b)
}

func (f FDComm) Read(b []byte) (int, error){
	return syscall.Read(f.FD,b)
}

type Client struct {
	FD          int
	ReadBuffer  []byte
	WriteBuffer []byte
}