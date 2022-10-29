package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sncp/internal/count"
	"sync"
	"syscall"
	"time"
)

const (
	minNum = 0
	maxNum = 9
)

var (
	fServerPort int
	// fProccessNum int
	fThreadNum int
)

func main() {
	flag.IntVar(&fServerPort, "port", 0, "Set a port for server connection")
	// flag.IntVar(&fProccessNum, "cp", 0, "Number of child-processes needs")
	flag.IntVar(&fThreadNum, "t", 0, "Number of threads needs")
	flag.Parse()

	if errs := validateParameters(); len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		os.Exit(5)
	}

	counter := count.New(minNum, maxNum)
	signal := make(chan syscall.Signal)
	wg := &sync.WaitGroup{}

	wg.Add(fThreadNum)
	for i := 1; i <= fThreadNum; i++ {
		go setCount(counter, wg, signal)
	}

	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		fmt.Println("Set triggered")
		signal <- syscall.SIGHUP
		close(signal)
	}()

	wg.Wait()
	fmt.Println("fThreadNum all done!")
	fmt.Println("Print total of each number")

	counter.Range(walk)
}

func validateParameters() []error {
	errs := []error{}

	if fServerPort == 0 {
		errs = append(errs, errors.New("please input a port number"))
	}

	// if fProccessNum == 0 {
	// 	errs = append(errs, errors.New("please input number of child-proccess needs"))
	// }

	if fThreadNum == 0 {
		errs = append(errs, errors.New("please input number of threads needs"))
	}

	return errs
}

func setCount(counter *count.Map, wg *sync.WaitGroup, signal <-chan syscall.Signal) {
	defer wg.Done()

	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-signal:
			return
		default:
			randNum := rand.Intn(maxNum-minNum+1) + minNum
			currentCount, _ := counter.Load(randNum)
			currentCount = currentCount.(int) + 1
			counter.Store(randNum, currentCount)
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
	}
}

func walk(key, value any) bool {
	fmt.Println("Key =", key, "count =", value.(int))
	return true
}
