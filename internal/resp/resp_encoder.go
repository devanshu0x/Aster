package resp

import (
	"errors"
	"fmt"
)

func encodeSimpleString(v *RESPValue) ([]byte,error){
	if v==nil || v.Type!=RESPSimpleString{
		return nil,errors.New("Invalid encoding type")
	}

	return []byte(fmt.Sprintf("+%s\r\n",v.Value)),nil
}

func encodeInteger(v *RESPValue) ([]byte,error){
	if v==nil || v.Type!=RESPInteger{
		return nil,errors.New("Invalid encoding type")
	}

	return []byte(fmt.Sprintf(":%d\r\n",v.Value)),nil
	
}

func encodeBulkString(v *RESPValue) ([]byte,error){
	if v==nil || v.Type!=RESPBulkString{
		return nil,errors.New("Invalid encoding type")
	}
	if v.Value==nil{
		return []byte("$-1\r\n"),nil
	}
	val,ok:=v.Value.(string)
	if !ok{
		return nil,errors.New("resp value is not string")
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n",len(val),val)),nil
	
}

func encodeArray(v *RESPValue) ([]byte,error){
	if v==nil || v.Type!=RESPArray{
		return nil,errors.New("Invalid encoding type")
	}
	if v.Value==nil{
		return []byte("*-1\r\n"),nil
	}

	arr,ok:=v.Value.([]*RESPValue)
	if !ok{
		return nil,errors.New("resp value is not array")
	}

	encoding:=[]byte(fmt.Sprintf("*%d\r\n",len(arr)))
	for _,ele:=range arr{
		val,err:=Encode(ele)
		if err!=nil{
			return nil,err
		}
		encoding=append(encoding, val...)
	}

	return encoding,nil
}

func encodeError(v *RESPValue) ([]byte,error){
	if v==nil || v.Type!=RESPError{
		return nil,errors.New("Invalid encoding type")
	}

	return []byte(fmt.Sprintf("-%s\r\n",v.Value)),nil
}


func Encode(v *RESPValue) ([]byte,error){

	switch v.Type{
	case RESPArray:
		return encodeArray(v)
	case RESPBulkString:
		return encodeBulkString(v)
	case RESPInteger:
		return encodeInteger(v)
	case RESPError:
		return encodeError(v)
	case RESPSimpleString:
		return encodeSimpleString(v)
	default:
		return nil, errors.New("Invalid RESP type")					
	}

}