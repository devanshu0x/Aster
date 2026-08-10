package command

import (
	"strings"

	"github.com/devanshu0x/Aster/internal/resp"
)

func RESPError(msg string) (*resp.RESPValue){
	return &resp.RESPValue{
		Type: resp.RESPError,
		Value: msg,
	}
}

func Dispatch(v *resp.RESPValue)(respVal *resp.RESPValue,is_mutated bool){
	if v==nil{
		return RESPError("NIL pointer wtf!"),false
	}
	if v.Type!=resp.RESPArray {
		return RESPError("ERR invalid command"),false
	}

	arr,ok:=v.Value.([]*resp.RESPValue)
	if !ok{
		return RESPError("ERR invalid command"),false
	}
	
	if len(arr)==0{
		return RESPError("ERR invalid command"),false
	}

	if arr[0].Type!=resp.RESPBulkString{
		return RESPError("ERR invalid command"),false
	}

	cmd,ok:=arr[0].Value.(string)
	if !ok{
		return RESPError("ERR invalid command"),false
	}
	
	cmd=strings.ToUpper(cmd)

	switch cmd{
	case "PING":
		return cmdPING(arr[1:]),false
	case "SET":
		return cmdSET(arr[1:])
	case "GET":
		return cmdGET(arr[1:]),false	
	case "TTL":		
		return cmdTTL(arr[1:]),false
	case "EXPIRE":
		return cmdEXPIRE(arr[1:])
	case "DEL":
		return cmdDEL(arr[1:])	
	case "SAVE":
		return cmdSAVE(),false		
	default:
		return RESPError("Unregisted command"),false	
	}


}