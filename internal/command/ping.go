package command

import "github.com/devanshu0x/Aster/internal/resp"

func cmdPing(argArr []*resp.RESPValue) *resp.RESPValue{
	if len(argArr)>2{
		return RESPError("More than one argument not allowed")
	}
	if len(argArr)==1{
		return &resp.RESPValue{
			Type: resp.RESPBulkString,
			Value: argArr[0],
		}
	}

	return &resp.RESPValue{
		Type: resp.RESPSimpleString,
		Value: "PONG",
	}
}