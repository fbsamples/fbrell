// Package stathat implements a stathat backend for go.stats.
package stathat

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/facebookgo/jsonpipe"
	"github.com/facebookgo/muster"
)

type countStat struct {
	Name  string `json:"stat"`
	Count int    `json:"count"`
}

type valueStat struct {
	Name  string  `json:"stat"`
	Value float64 `json:"value"`
}

type batchResponse struct {
	Status   int    `json:"status"`
	Message  string `json:"msg"`
	Multiple int    `json:"multiple"`
}

type Client struct {
	Key                 string        // your StatHat EZ Key
	Debug               bool          // enable logging of stat calls
	BatchTimeout        time.Duration // timeout for batching stats
	MaxBatchSize        uint          // max items in a batch
	PendingWorkCapacity uint          // buffer size until we begin blocking
	Transport           http.RoundTripper
	muster              muster.Client
}

// Start the background goroutine for handling the actual HTTP requests.
func (c *Client) Start() error {
	c.muster.MaxBatchSize = c.MaxBatchSize
	c.muster.BatchTimeout = c.BatchTimeout
	c.muster.PendingWorkCapacity = c.PendingWorkCapacity
	c.muster.BatchMaker = func() muster.Batch {
		return &batch{Client: c, Key: c.Key}
	}
	return c.muster.Start()
}

// Stop the background goroutine.
func (c *Client) Stop() error {
	return c.muster.Stop()
}

// Increment the named counter by the given value.
func (c *Client) Count(name string, count int) {
	c.muster.Work <- countStat{Name: name, Count: count}
}

// Record an instance of the given value for the given name.
func (c *Client) Record(name string, value float64) {
	c.muster.Work <- valueStat{Name: name, Value: value}
}

// Special case increment by 1 for the named counter.
func (c *Client) Inc(name string) {
	c.Count(name, 1)
}

// A Flag configured Client instance.
func ClientFlag(name string) *Client {
	c := &Client{}
	flag.StringVar(&c.Key, name+".key", "", name+" ezkey")
	flag.BoolVar(&c.Debug, name+".debug", false, name+" debug logging")
	flag.DurationVar(
		&c.BatchTimeout,
		name+".batch-timeout",
		10*time.Second,
		name+" amount of time to aggregate a batch",
	)
	flag.UintVar(
		&c.MaxBatchSize,
		name+".max-batch-size",
		1000,
		name+" maximum number of items in a batch",
	)
	flag.UintVar(
		&c.PendingWorkCapacity,
		name+".pending-work-capacity",
		10000,
		name+" pending work capacity",
	)
	return c
}

type batch struct {
	Client *Client       `json:"-"`
	Key    string        `json:"ezkey"`
	Data   []interface{} `json:"data"` // countStat or valueStat
}

func (b *batch) Add(stat interface{}) {
	if b.Client.Debug {
		if cs, ok := stat.(countStat); ok {
			log.Printf("stathat: Count(%s, %d)", cs.Name, cs.Count)
		}
		if vs, ok := stat.(valueStat); ok {
			log.Printf("stathat: Value(%s, %f)", vs.Name, vs.Value)
		}
	}
	b.Data = append(b.Data, stat)
}

func (b *batch) Fire(notifier muster.Notifier) {
	defer notifier.Done()
	if err := b.fire(); err != nil {
		log.Println(err)
	}
}

func (b *batch) fire() error {
	if b.Client.Debug {
		log.Printf("stathat: sending batch with %d items", len(b.Data))
	}

	const url = "http://api.stathat.com/ez"
	req, err := http.NewRequest("POST", url, jsonpipe.Encode(b))
	if err != nil {
		return fmt.Errorf("stathat: error creating http request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := b.Client.Transport.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("stathat: %s", err)
	}
	defer resp.Body.Close()
	var br batchResponse
	err = json.NewDecoder(resp.Body).Decode(&br)
	if err != nil {
		return fmt.Errorf("stathat: error decoding response: %s", err)
	}
	if br.Status != 200 {
		return fmt.Errorf("stathat: api error: %+v", &br)
	} else if b.Client.Debug {
		log.Printf("stathat: api response: %+v", &br)
	}
	return nil
}
