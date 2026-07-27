package command

import (
	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdDEL(argArr []*resp.RESPValue) *resp.RESPValue{
	if len(argArr)==0{
		return RESPError("Err invalid number of argument from 'del' command")
	}

	count:=0;
	for _,val:=range argArr{
		key,ok:=val.Value.(string)
		if !ok{
			return RESPError("Err expected string as a key")
		}
		if store.Del(key){
			count++
		}
	}

	return &resp.RESPValue{
		Type: resp.RESPInteger,
		Value: count,
	}
}