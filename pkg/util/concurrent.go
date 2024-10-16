package util

import (
	"sync"
)

func Generator(done <-chan bool, data ...string) <-chan string {
	readStream := make(chan string)
	go func() {
		defer close(readStream)
		for _, s := range data {
			select {
			case <-done:
				return
			case readStream <- s:
			}
		}
	}()

	return readStream
}

func FanIn(done <-chan bool, channels ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	multiplexedStream := make(chan string)

	multiplex := func(c <-chan string) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- i:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}
