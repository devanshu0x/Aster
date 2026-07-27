package command

import (
	"strconv"
	"strings"

	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdSET(argArr []*resp.RESPValue) *resp.RESPValue{
	if len(argArr)<=1{
		return RESPError("Err wrong number of argument for 'set' command")
	}

	key,ok:=argArr[0].Value.(string)
	if !ok{
		return RESPError("Err expected string as a key")
	}

	val,ok:=argArr[1].Value.(string)
	if !ok{
		return RESPError("Err expected string as a value")
	}

	exDurationMs:=int64(-1)

	for i:=2;i<len(argArr);i++{
		arg,ok:=argArr[i].Value.(string)
		if !ok{
			return RESPError("Err expected string as an argument")
		}
		arg=strings.ToUpper(arg)
		switch(arg){
		case "EX":
			i++
			if i==len(argArr){
				return RESPError("Err invalid syntax")
			}
			durationString,ok:=argArr[i].Value.(string)
			if !ok{
				return RESPError("Err expected string as arg value")
			}
			exDurationSec,err:=strconv.ParseInt(durationString,10,64)
			if err!=nil{
				return RESPError("Err value is not integer or out of range")
			}
			exDurationMs=exDurationSec*1000
		default:
			return RESPError("Err invalid syntax")	
		}

	}

	store.Put(key,store.NewObject(val,exDurationMs,store.StringObject))
	return &resp.RESPValue{
			Type:resp.RESPSimpleString ,
			Value: "OK",
		}
	
}