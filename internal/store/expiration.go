// package store

// import (
// 	"log"
// 	"time"
// )

// func expireSample() float32{
// 	limit:=20
// 	expiredCount:=0

// 	// assuming iteration in golang hash table is randomized

// 	for key,obj:= range store{
// 		if obj.ExpiresAt==-1{
// 			continue
// 		}
// 		// Current obj do have some ttl value not -1
// 		limit--
// 		if obj.ExpiresAt<=time.Now().UnixMilli(){
// 			delete(store,key)
// 			expiredCount++;
// 		}

// 		if limit==0{
// 			break
// 		}
// 	}
// 	sampled:=20-limit
// 	if sampled==0{
// 		return 0
// 	}
// 	return  float32(expiredCount)/float32(sampled)
// }


// // active deletion: delete all the expired key
// func DeleteExpiredKeys(){
// 	for{
// 		// if the sample had less than 25% keys expired
// 		frac:=expireSample()

// 		if frac<0.25{
// 			break
// 		}
// 	}

// 	log.Println("Deleted expired but undeleted keys. Total keys left: ",len(store))
// }