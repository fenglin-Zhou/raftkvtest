package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"
	"unsafe"

	v3 "go.etcd.io/etcd/client/v3"

	pb "github.com/cheggaaa/pb/v3"
)

var (
	wg     sync.WaitGroup
	result chan Result
	done   chan int
	putnum = flag.Int("r", 10000, "input put number")
	clinum = flag.Int("c", 1, "input cli number")
	Timer  *time.Timer
)

func main() {
	flag.Parse()
	bar := pb.New(*putnum)
	bar.Start()
	result = make(chan Result, 16)
	done = make(chan int, 10)
	for i := 0; i < *clinum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rootContext := context.Background()
			cli, err := v3.New(v3.Config{
				Endpoints:   []string{"127.0.0.1:23793"},
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
			for bar.Current() < int64(*putnum) {
				k := RandStr(5)
				v := RandStr(5)
				t1 := time.Now()
				_, err := kvc.Put(rootContext, k, v)
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
	Timer = time.NewTimer(time.Millisecond * 1000)
	go func() {
		filepath := "./test1.txt"
		file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("文件打开失败", err)
		}
		//及时关闭file句柄
		defer file.Close()
		//写入文件时，使用带缓存的 *Writer
		write := bufio.NewWriter(file)
		total := bar.Current()
		loop := true
		for loop {
			select {
			case <-Timer.C:
				write.WriteString(fmt.Sprintf("%d\n", bar.Current()-total))
				total = bar.Current()
				Timer.Reset(time.Millisecond * 1000)
			case <-done:
				Timer.Stop()
				write.Flush()
				loop = false
			}
		}
		Timer.Stop()
	}()
	go Report()
	wg.Wait()
	done <- 1
	done <- 1
	time.Sleep(time.Second * 5)
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
			if i == 200 {
				i = 0
				newlats := lats[20:180]
				for _, lat := range newlats {
					total = total + lat
				}
				write.WriteString(fmt.Sprintf("%f\n", total/float64(160)))
				total = 0.0
				lats = nil
			}
		case <-done:
			write.Flush()
			loop = false
		}
	}
}
