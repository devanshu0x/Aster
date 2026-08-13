package command

import (
	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdINCR(argArr []*resp.RESPValue) (*resp.RESPValue,bool){
	if len(argArr)!=1{
		return RESPError("Err wrong number of argument for 'incr' command"),false
	}

	key,ok:=argArr[0].Value.(string)
	if !ok{
		return RESPError("Err expected string as a key"),false
	}

	if val,ok:=store.Incr(key);ok{
		return &resp.RESPValue{
			Type: resp.RESPInteger,
			Value: val,
		},true
	}else{
		return RESPError("Err value is not integer"),false
	}
}