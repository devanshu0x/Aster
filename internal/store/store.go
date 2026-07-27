package store

import (
	"time"
)

var store map[string] *Obj

type ObjType string

const(
	StringObject ObjType="string"
	HashObject ObjType="hash"
	ListObject ObjType="list"
)

type Obj struct{
	Value interface{}
	ExpiresAt int64
	Type ObjType
}

func init(){
	store= make(map[string] *Obj)
}

func NewObject(value interface{},durationMs int64, objType ObjType) *Obj{
	 expiresAt:= int64(-1)
	if durationMs>0{
		expiresAt=time.Now().UnixMilli()+durationMs
	}

	return &Obj{
		Value: value,
		ExpiresAt: expiresAt,
		Type: objType,
	}
}

func Put(k string,obj *Obj){
	store[k]=obj
}

func Get(k string) *Obj{
	return store[k]
}

func Del(k string) bool{
	obj,ok:=store[k]
	if !ok{
		return false
	}
	if obj.ExpiresAt!=-1 && obj.ExpiresAt<=time.Now().UnixMilli(){
		delete(store,k)
		return false
	}

	delete(store,k)
	return true
}

func Expire(k string, expInMilli int64) bool{
	if _,ok:=store[k];!ok{
		return false
	}

	obj:=store[k]
	if obj.ExpiresAt!=-1 && obj.ExpiresAt<=time.Now().UnixMilli(){
		return false
	}

	expTime:=time.Now().UnixMilli()+expInMilli

	obj.ExpiresAt=expTime

	return true
}