package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	pb "github.com/cheggaaa/pb/v3"
	"github.com/thanhpk/randstr"
	v3 "go.etcd.io/etcd/client/v3"
)

var (
	wg       sync.WaitGroup
	result   chan Result
	done     chan int
	putnum   = flag.Int("r", 10000, "input put number")
	clinum   = flag.Int("c", 1, "input cli number")
	endpoint = flag.String("end", "127.0.0.1:2379", "input ip port")
	Timer    *time.Timer
)

type kv struct {
	k string
	v string
}

func main() {
	flag.Parse()
	bar := pb.New(*putnum)
	bar.SetRefreshRate(time.Second)
	bar.Start()
	file, _ := os.Create("./bar.txt")
	defer file.Close()
	bar.SetWriter(file)
	result = make(chan Result, 100)
	req := make(chan kv, 100)
	done = make(chan int, 10)
	for i := 0; i < *clinum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rootContext := context.Background()
			cli, err := v3.New(v3.Config{
				Endpoints:   []string{*endpoint},
				DialTimeout: 2 * time.Second,
			})
			fmt.Println("create cli")
			if cli == nil || err == context.DeadlineExceeded {
				// handle errors
				fmt.Println(err)
				panic("invalid connection!")
			}
			// 客户端断开连接
			defer cli.Close()
			// 初始化 kv
			kvc := v3.NewKV(cli)
			for op := range req {
				t1 := time.Now()
				_, err := kvc.Put(rootContext, op.k, op.v)
				if err != nil {
				}
				// if bar.Current() > 100 {
				// 	result <- Result{start: t1, end: time.Now()}
				// }
				result <- Result{start: t1, end: time.Now()}
				bar.Increment()
			}
		}()
	}
	go func() {
		for i := 0; i < *putnum; i++ {
			req <- kv{randstr.String(8), randstr.String(128)}
		}
		close(req)
	}()
	go Report()
	wg.Wait()
	done <- 1
	time.Sleep(time.Second * 5)
}

type Result struct {
	start time.Time
	end   time.Time
}

func Report() {
	filepath := "./test.txt"
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	i := 0
	var total float64
	var lats []float64
	total = 0.0
	loop := true
	for loop {
		select {
		case res := <-result:
			lats = append(lats, res.end.Sub(res.start).Seconds())
			sort.Float64s(lats)
			i = i + 1
			if i == 400 {
				i = 0
				newlats := lats[40:360]
				for _, lat := range newlats {
					total = total + lat
				}
				write.WriteString(fmt.Sprintf("%f\n", total/float64(320)))
				total = 0.0
				lats = nil
			}
		case <-done:
			write.Flush()
			loop = false
		}
	}
}
