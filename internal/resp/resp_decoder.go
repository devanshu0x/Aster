/*
Package resp implements RESP (Redis Serialization Protocol).

RESP is the wire protocol used by Redis clients and servers to exchange
commands and responses over TCP.

Implementing RESP allows Aster to communicate with existing Redis clients
and client libraries without requiring any custom protocol.

*/

package resp

import (
	"bytes"
	"errors"
	"strconv"
)



func decodeSimpleString(d []byte)(v *RESPValue,n int,done bool, err error){
	idx:=bytes.Index(d,[]byte("\r\n"))
	if idx==-1{
		return nil,0,false,nil
	}
	return &RESPValue{
		Type: RESPSimpleString,
		Value: string(d[:idx]),
	},idx+2,true,nil
}

func decodeInt(d []byte)(v *RESPValue,n int,done bool, err error){
	idx:=bytes.Index(d,[]byte("\r\n"))
	if idx==-1{
		return nil,0,false,nil
	}
	val,err:=strconv.Atoi(string(d[:idx]))
	if err!=nil{
		return nil,0,false,err
	}

	return &RESPValue{
		Type: RESPInteger,
		Value: val,
	},idx+2,true,nil
}

func decodeBulkString(d []byte)(v *RESPValue,n int,done bool, err error){
	strLen,n,done,err:=decodeInt(d)
	if !done || err!=nil{
		return nil,0,done,err
	}

	// assert strlen value is int 
	length:= strLen.Value.(int)

	// strLen==-1 is null bulk string : $-1\r\n
	if length==-1{
		return nil,n,true,nil
		// wait what should i return in case of null bulk string? Maybe nil for now
	} 
	// -ve length bulk string is not allowed
	if length < -1{
		return nil,0,false,errors.New("negative bulk string length")
	}

	// byte stream is not complete for parsing
	if length+n+2>len(d){
		return nil,0,false,nil
	}

	if !bytes.HasPrefix(d[n+length:],[]byte("\r\n")) {
		return nil,0,false, errors.New("Bulk string not correctly encoded")
	}

	return &RESPValue{
		Type: RESPBulkString,
		Value: string(d[n:n+length]),
	},n+length+2,true,nil

}

func decodeArray(d []byte)(v *RESPValue,n int,done bool, err error){
	arrLen,n,done,err:=decodeInt(d)
	if !done || err!=nil{
		return nil,0,done,err
	}
	// assert length to be int
	length:=arrLen.Value.(int)
	// null array
	if length== -1{
		return nil,n,true,nil
	}

	// reject -ve length array
	if length < -1{
		return nil,0,false,errors.New("negative array length")
	}

	arr:=make([]*RESPValue,length)
	
	for i:= range length{
		val,bytesR,done,err:=Decode(d[n:])
		if !done || err!=nil{
			return nil,0,done,err
		}
		n+=bytesR
		arr[i]=val
	}

	return &RESPValue{
		Type: RESPArray,
		Value: arr,
	},n,true,nil
}

func decodeError(d []byte) (v *RESPValue,n int,done bool, err error){
	idx:=bytes.Index(d,[]byte("\r\n"))
	if idx==-1{
		return nil,0,false,nil
	}
	return &RESPValue{
		Type: RESPError,
		Value: string(d[:idx]),
	},idx+2,true,nil
}



// Redis client requests are always encoded as RESP arrays.
// Aster uses this property to determine whether a complete
// command has been received over TCP.
func Decode(d []byte) (v *RESPValue,n int, done bool, err error){
	if len(d)==0{
		return nil,0,false,nil
	}
	
	switch d[0]{
	case '+':
		v,n,done,err=decodeSimpleString(d[1:])
	case ':':
		v,n,done,err=decodeInt(d[1:])
	case '*':
		v,n,done,err=decodeArray(d[1:])
	case '$':
		v,n,done,err=decodeBulkString(d[1:])
	case '-':
		v,n,done,err=decodeError(d[1:])
	default:
		return nil,0,false,errors.New("Unsupported type")				
	}

	if done{
		// first byte that we read in switch statement
		n++
	}
	return v,n,done,err

}