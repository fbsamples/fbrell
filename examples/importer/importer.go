package main

import (
	"flag"
	"fmt"
	"github.com/daaku/go.flagconfig"
	"github.com/daaku/rell/examples"
	"launchpad.net/goamz/s3"
	"net/http"
	"runtime"
	"sync"
)

func safeLoad(key string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Failed Example %s: %s\n", key, r)
		}
	}()
	_, err := examples.Load("mu", fmt.Sprintf("/saved/%s", key))
	if err != nil {
		panic(err)
	}
}

func get(markers chan string, wg *sync.WaitGroup, bucket *s3.Bucket) {
	for marker := range markers {
		res, err := bucket.List("", "", marker, 1000)
		if err != nil {
			fmt.Printf("Failed List %s: %s\n", marker, err)
			go func() { markers <- marker }()
			return
		}
		if res.IsTruncated {
			wg.Add(1)
			go func() {
				markers <- res.Contents[len(res.Contents)-1].Key
			}()
		}
		for _, key := range res.Contents {
			safeLoad(key.Key)
		}
		wg.Done()
	}
}

func main() {
	flag.Parse()
	flagconfig.Parse()
	runtime.GOMAXPROCS(4)
	http.DefaultClient.Transport = &http.Transport{
		DisableKeepAlives: true,
	}
	markers := make(chan string)
	wg := new(sync.WaitGroup)
	bucket := examples.Bucket()
	go get(markers, wg, bucket)
	go get(markers, wg, bucket)
	go get(markers, wg, bucket)
	go get(markers, wg, bucket)
	wg.Add(1)
	markers <- ""
	wg.Wait()
	fmt.Println("Finished.")
}
