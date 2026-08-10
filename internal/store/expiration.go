package store

import (
	"log"
	"math/rand"
	"time"
)

func expireSample() float64 {
	limit:=min(20,store.ht[0].used)
	expiredCount:=0
	sampled:=0

	if limit<=0{
		return 0
	}

	for sampled<limit{
		idx:=rand.Intn(len(store.ht[0].buckets))

		curr:=store.ht[0].buckets[idx]

		for curr!=nil && sampled<limit{
			sampled++
			obj:=curr.Obj
			next:=curr.Next
			if obj.ExpiresAt!=-1 && obj.ExpiresAt<=time.Now().UnixMilli(){
				store.ht[0].deleteEntry(curr.Key)
				expiredCount++
			}

			curr=next
		}
	}

	return float64(expiredCount)/float64(sampled)
}


// active deletion: delete all the expired key
func DeleteExpiredKeys(){
	// skip active deletion during rehashing
	if store.isRehashing(){
		return
	}
	deleted:=false
	for{
		// if the sample had less than 25% keys expired
		frac:=expireSample()
		if frac!=0{
			deleted=true
		}
		if frac<0.25{
			break
		}
	}

	if deleted{
		log.Println("Deleted expired but undeleted keys. Total keys left: ",store.ht[0].used)
	}
}