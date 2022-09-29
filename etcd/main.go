package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	randstr "github.com/fenglin-Zhou/raftkvtest/randstr"
	v3 "go.etcd.io/etcd/client/v3"
)

var (
	wg sync.WaitGroup
)

func main() {
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go work()
	}
	wg.Wait()
}
func work() {
	defer wg.Done()
	rootContext := context.Background()
	cli, err := v3.New(v3.Config{
		Endpoints:   []string{"127.0.0.1:23791"},
		DialTimeout: 2 * time.Second,
	})
	if cli == nil || err == context.DeadlineExceeded {
		// handle errors
		fmt.Println(err)
		panic("invalid connection!")
	}
	// 客户端断开连接
	defer cli.Close()
	// 初始化 kv
	kvc := v3.NewKV(cli)

	for j := 0; j < 30; j++ {
		t1 := time.Now()
		for i := 0; i < 1000; i++ {
			k := randstr.RandStr(5)
			v := randstr.RandStr(5)
			_, err := kvc.Put(rootContext, k, v)
			if err != nil {
			}
		}
		t2 := time.Now()
		fmt.Println("total time: ", t2.Sub(t1))
	}
}
