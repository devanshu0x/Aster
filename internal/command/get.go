package command

import (
	"time"

	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdGET(argArr []*resp.RESPValue) *resp.RESPValue {
	if len(argArr)!=1{
		return RESPError("Err wrong number of argument for 'get' command")
	}

	key,ok:=argArr[0].Value.(string)
	if !ok{
		return RESPError("Err expected string as a key")
	}

	valObj:=store.Get(key)

	if valObj==nil{
		return &resp.RESPValue{
			Type: resp.RESPBulkString,
			Value: nil,
		}
	}

	if  valObj.ExpiresAt!=-1 && valObj.ExpiresAt<=time.Now().Unix(){
		return &resp.RESPValue{
			Type: resp.RESPBulkString,
			Value: nil,
		}
	}

	return &resp.RESPValue{
		Type: resp.RESPBulkString,
		Value: valObj.Value,
	}
}