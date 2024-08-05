package crawler

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

const chanCapacity = 100000

func (c *Crawler) ParseParallel() {
	ctx, cancel := context.WithCancel(context.Background())

	numCPU := runtime.NumCPU()
	urlChan := make(chan string, chanCapacity)

	fmt.Println("Num of workers: ", numCPU)

	urlChan <- c.targetURL

	for i := 0; i < numCPU; i++ {
		c.wg.Add(1)

		worker := NewWorker(i+1, urlChan, c)
		go worker.run(ctx)
	}

	// wait for filling chan by some worker
	time.Sleep(1 * time.Second)

	var l int

	// if chan will become empty just close all workers and out of parsing
	for {
		l = len(urlChan)
		if l == 0 {
			cancel()
			break
		}

		time.Sleep(1 * time.Second)

	}

	c.wg.Wait()
}

type worker struct {
	id      int
	ch      chan string
	crawler *Crawler
}

func NewWorker(id int, ch chan string, c *Crawler) *worker {
	return &worker{
		id,
		ch,
		c,
	}
}

func (w *worker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d exiting\n", w.id)
			w.crawler.wg.Done()
			return
		case link := <-w.ch:
			if w.crawler.checkAndAdd(link) {
				continue
			}

			urls, err := w.crawler.urls(link)
			if err != nil {
				fmt.Println(err)
			}

			var i int
			for i = range urls {
				w.ch <- urls[i]
			}
		}
	}
}
