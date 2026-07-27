package command

import (
	"strconv"

	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdEXPIRE(argArr []*resp.RESPValue) *resp.RESPValue{
	if len(argArr)!=2{
		return RESPError("Err invalid number of argument for 'expire' command")
	}

	key,ok:=argArr[0].Value.(string)
	if !ok{
		return RESPError("Err expected string as a key")
	}

	expTimeString,ok:=argArr[1].Value.(string)
	if !ok{
		return RESPError("Err expected string as a argument value")
	}

	expTimeInSecond,err:=strconv.ParseInt(expTimeString,10,64)
	if err!=nil{
		return RESPError("Err expected expiration time to be integer")
	}
	count:=0
	if store.Expire(key,expTimeInSecond*1000){
		count++
	}

	return &resp.RESPValue{
		Type: resp.RESPInteger,
		Value: count,
	}


}