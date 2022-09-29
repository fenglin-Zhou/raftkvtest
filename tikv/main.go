package main

import (
	"fmt"
	"sync"
	"time"

	randstr "github.com/fenglin-Zhou/raftkvtest/randstr"
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
			k := randstr.RandStr(5)
			v := randstr.RandStr(5)
			err := cli.Put([]byte(k), []byte(v))
			if err != nil {
				panic(err)
			}
		}
		t2 := time.Now()
		fmt.Println("total time: ", t2.Sub(t1))
	}
}
