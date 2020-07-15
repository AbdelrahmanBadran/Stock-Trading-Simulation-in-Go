package main

import (	
	"container/list"
	"sync"
	"testing"
	"time"		
	"runtime"	
)

var result time.Duration

func benchmarkStock(b *testing.B) {		
	runtime.GOMAXPROCS(1) //runtime.NumCPU()		

	var r time.Duration
	startTime := time.Now()

	for n := 0; n < b.N; n++ {		
		SIMSIZE := 25

		var stocksTrack safeStockTrack
		var stocksOwned safeStocks
		for i := 0; i < SIMSIZE; i++ { stocksTrack.valList[i] = list.New() }	

		updateChan := make(chan string, SIMSIZE)
		var wg sync.WaitGroup
		
		r = trackingChanges(updateChan, &stocksTrack, &stocksOwned, &wg, &SIMSIZE, &startTime)				
		wg.Wait()
	}	
	result = r		
}

func BenchmarkStock(b *testing.B) { benchmarkStock(b) }

// 2	 512260450 ns/op	168006956 B/op	 4000038 allocs/op