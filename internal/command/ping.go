package command

import "github.com/devanshu0x/Aster/internal/resp"

func cmdPING(argArr []*resp.RESPValue) *resp.RESPValue{
	if len(argArr)>2{
		return RESPError("Err wrong number of argument for 'ping' command")
	}
	if len(argArr)==1{
		return &resp.RESPValue{
			Type: resp.RESPBulkString,
			Value: argArr[0].Value,
		}
	}

	return &resp.RESPValue{
		Type: resp.RESPSimpleString,
		Value: "PONG",
	}
}