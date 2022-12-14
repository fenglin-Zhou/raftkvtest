package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"unsafe"

	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/store/tikv"
)

var (
	wg sync.WaitGroup
)

func main() {
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go Work()
	}
	wg.Wait()
}

func Work() {
	cli, err := tikv.NewRawKVClient([]string{"127.0.0.1:2379"}, config.Security{})
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	for j := 0; j < 30; j++ {
		t1 := time.Now()
		for i := 0; i < 1000; i++ {
			k := RandStr(5)
			v := RandStr(5)
			err := cli.Put([]byte(k), []byte(v))
			if err != nil {
				// panic(err)
			}
		}
		t2 := time.Now()
		fmt.Println("total time: ", t2.Sub(t1))
	}
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

func RandStr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}
