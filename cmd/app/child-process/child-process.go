package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sncp/internal/count"
	"sncp/internal/intranet"
	"syscall"
)

const (
	minNum = 0
	maxNum = 9
)

var (
	fServerPort int
	fThreadNum  int
	counter     = count.New(minNum, maxNum)
	tcpClient   intranet.Client
)

func main() {
	flag.IntVar(&fServerPort, "port", 0, "Set a port for server connection")
	flag.IntVar(&fThreadNum, "t", 0, "Number of threads needs")
	flag.Parse()
	if errs := validateParameters(); len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		os.Exit(5)
	}

	var err error
	tcpClient, err = intranet.NewClient(fServerPort)
	if err != nil {
		fmt.Println("New client error: ", err)
		os.Exit(5)
	}

	tcpClient.SetRecvCallBack(readRecvMap)
	tcpClient.Run()

	if err != nil {
		fmt.Println("Create tcp client error : ", err.Error())
	}

	go counter.Start(fThreadNum)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	// go func() {
	// 	for range time.Tick(1 * time.Second) {
	// 		time := time.Now().Add(-1 * time.Second)
	// 		counter.Pause()

	// 		pkg := &intranet.CountPackage{}
	// 		pkg.Time = time.Unix()
	// 		pkg.Data = [10]int{}

	// 		for i := counter.GetMin(); i <= counter.GetMax(); i++ {
	// 			value, _ := counter.Load(i)
	// 			pkg.Data[i] = value.(int)
	// 		}
	// 		buf := bytes.NewBuffer(nil)
	// 		_ = gob.NewEncoder(buf).Encode(&pkg)

	// 		err := tcpClient.Send(buf.Bytes())
	// 		if err != nil {
	// 			fmt.Println("Send data error ", err.Error())
	// 		}

	// 		counter.Reset()
	// 		counter.Resume()
	// 	}
	// }()
	errorCh := tcpClient.GetErrorChannel()
	for {
		select {
		case err := <-errorCh:
			fmt.Println("Client error: ", err)
		case <-sig:
			counter.Done()
			return
		}
	}
}

func readRecvMap(data []byte, errCh chan error) {
	pkg := &intranet.CountPackage{}
	buf := bytes.NewBuffer(data)
	err := gob.NewDecoder(buf).Decode(&pkg)
	if err != nil {
		errCh <- fmt.Errorf("parse received data error: %s", err.Error())
		return
	}

	for i := counter.GetMin(); i <= counter.GetMax(); i++ {
		value, _ := counter.LoadMap(i)
		pkg.Data[i] = value.(int)
	}

	counter.Reset()
	counter.Resume()

	sendBuf := bytes.NewBuffer(nil)
	_ = gob.NewEncoder(sendBuf).Encode(&pkg)

	tcpClient.Send(sendBuf.Bytes())
}

func validateParameters() []error {
	errs := []error{}

	if fServerPort == 0 {
		errs = append(errs, errors.New("please input a port number"))
	}

	if fThreadNum == 0 {
		errs = append(errs, errors.New("please input number of threads needs"))
	}

	return errs
}
