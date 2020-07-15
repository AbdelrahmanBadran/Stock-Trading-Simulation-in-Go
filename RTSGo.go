package main

import (
	"bufio"
	"container/list"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	// "github.com/pkg/profile"
)

type safeStocks struct {
	value [25]float64
	mutex [25]sync.Mutex
}

type safeStockTrack struct {
	valList [25]*list.List
	mutex   [25]sync.Mutex
}

type structStock struct {
	name [25]string
}

func main() {
	runtime.GOMAXPROCS(1)
	// defer profile.Start(profile.MemProfile).Stop() //TraceProfile

	startTime := time.Now()
	SIMSIZE := 25

	path := "stocks.txt"
	file, err := os.Create(path)
	check(err)

	var stockValues safeStocks
	var stockNames structStock
	stockNames.name = [25]string{"AAP", "MSF", "IBM", "ORC", "FCB", "GOL", "XRX", "STX", "NOR", "CAD",
		"AMD", "INT", "SAS", "ABA", "TEN", "UBR", "GRB", "VER", "AEG", "AIA",
		"AIG", "ANT", "ATT", "BTT", "CCC"}
	var fileMutex sync.Mutex
	updateStocks(&stockNames, &stockValues, &fileMutex, file, &startTime, &SIMSIZE)

	var stocksTrack safeStockTrack
	var ownedStocks safeStocks
	for i := 0; i < SIMSIZE; i++ {
		stocksTrack.valList[i] = list.New()
	}

	var wg sync.WaitGroup
	updateChan := make(chan string, SIMSIZE)
	trackingChanges(updateChan, &stocksTrack, &ownedStocks, &wg, &SIMSIZE, &startTime)

	wg.Wait()
	fmt.Printf("SYSTEM TERMINATED AFTER: %s\n", time.Since(startTime))
}

func trackingChanges(updateChan chan string, stocksTrack *safeStockTrack, ownedStocks *safeStocks,
	wg *sync.WaitGroup, simSize *int, startTime *time.Time) time.Duration {

	path := "stocks.txt"
	go func() {
		file, errs := os.Open(path)
		check(errs)
		for{ //Comment for Bench
			_, err := file.Seek(0, 1)
			check(err)

			s := bufio.NewScanner(file) //IMP
			for s.Scan() {
				updateChan <- s.Text()
			}
		}
		// close(updateChan)
	}()

	TRACKSIZE := 10
	for i := 0; i < *simSize; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			for stockUpdate := range updateChan {

				// fmt.Printf("%v\n", stockUpdate) //COMMENTED

				splitStock := strings.Split(stockUpdate, " ")

				stockID, err := strconv.ParseInt(splitStock[0], 10, 64)
				check(err)
				stockName := splitStock[1]
				updatedValue, err := strconv.ParseFloat(splitStock[2], 64)
				check(err)

				stocksTrack.mutex[stockID].Lock()
				stockHistory := stocksTrack.valList[stockID]
				stockHistory.PushBack(updatedValue) //IMP

				if stockHistory.Len() > TRACKSIZE {
					e := stockHistory.Front()
					stockHistory.Remove(e)

					ownedStocks.mutex[stockID].Lock()
					buySellAlgorithm(&stockName, &updatedValue, stockHistory, &ownedStocks.value[stockID])
					ownedStocks.mutex[stockID].Unlock()
				}

				stocksTrack.mutex[stockID].Unlock()
			}
			// fmt.Printf("GoRoutine %v Exited\n", i)
		}(i)
	}
	return time.Since(*startTime)
}

func buySellAlgorithm(stockName *string, updatedValue *float64, stockHistory *list.List, ownedStock *float64) {

	stockValsum := float64(0.0)
	for e := stockHistory.Front(); e != nil; e = e.Next() {
		stockValsum += e.Value.(float64)
	}
	stockAvg := stockValsum / float64(stockHistory.Len())

	if *ownedStock == float64(0.0) {
		if stockAvg < float64(-1.0) && stockAvg > float64(-2.0) {
			fmt.Printf("BUYS : %s	U: %.3f		A: %.3f\n", *stockName, *updatedValue, stockAvg)
			*ownedStock = *updatedValue
		}
	} else {
		if (stockAvg - *ownedStock) > float64(3.0) {
			fmt.Printf("SELLS: %s	U: %.3f		A: %.3f		H: %.3f\n",
				*stockName, *updatedValue, stockAvg, *ownedStock)
			*ownedStock = float64(0.0)
		}
	}
}

func updateStocks(stockNames *structStock, stocks *safeStocks,
	fileMutex *sync.Mutex, file *os.File, startTime *time.Time, simSize *int) {

	for i := 0; i < *simSize; i++ {
		go func() {
			ticker := time.NewTicker(time.Millisecond)
			for ; true; <-ticker.C {
				stockID := rand.Intn(25)
				stocks.mutex[stockID].Lock()

				stocks.value[stockID] += rand.Float64()*(1 - -1) - 1

				fileMutex.Lock()
				saveStockUpdates(file, fmt.Sprintf("%d", stockID)+" "+stockNames.name[stockID]+
					" "+fmt.Sprintf("%.16f", stocks.value[stockID])+
					" "+fmt.Sprintf("%s", time.Since(*startTime))+"\n")
				fileMutex.Unlock()

				stocks.mutex[stockID].Unlock()
			}
		}()
	}
}

func saveStockUpdates(file *os.File, update string) { //IMP
	writer := bufio.NewWriter(file)
	writer.WriteString(update)
	writer.Flush()
}

func check(e error) {
	if e != nil {
		fmt.Printf("err: %s", e)
	}
}
