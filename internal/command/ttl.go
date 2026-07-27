package command

import (
	"time"

	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdTTL(argArr []*resp.RESPValue) *resp.RESPValue{
	if len(argArr)!=1{
		return RESPError("Err invalid number of argument for 'ttl' command")
	}

	key,ok:=argArr[0].Value.(string)
	if !ok{
		return RESPError("Err expected string as key")
	}

	obj:=store.Get(key)

	// if obj with corresponding key does not exists send -2
	if obj==nil{
		return &resp.RESPValue{
			Type: resp.RESPInteger,
			Value: -2,
		}
	}

	// if Exp not set return -1
	if obj.ExpiresAt==-1{
		return  &resp.RESPValue{
			Type: resp.RESPInteger,
			Value: -1,
		}
	}

	durationMs:=obj.ExpiresAt -time.Now().UnixMilli()

	// object has expired
	if durationMs<0{
		return &resp.RESPValue{
			Type: resp.RESPInteger,
			Value: -2,
		}
	}

	return &resp.RESPValue{
		Type: resp.RESPInteger,
		Value: (durationMs)/1000,
	}

}