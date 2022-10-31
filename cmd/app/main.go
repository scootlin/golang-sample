package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sncp/internal/intranet"
	"sync"
	"syscall"
	"time"
)

var (
	fProccessNum   int
	fProccessName  string
	fOutPutFile    string
	fThreadNum     int
	childPs        []*os.Process
	countMap       sync.Map
	totalCountJson = []byte("[")
)

type CountInfo struct {
	Count int
	Data  [10]int
}

type JsonData struct {
	Time   int64
	Counts map[int]int
}

func main() {
	// Set Paramaters
	flag.IntVar(&fProccessNum, "ps", 1, "Number of child-processes needs")
	flag.IntVar(&fThreadNum, "th", 1, "Number of threads needs")
	flag.StringVar(&fProccessName, "filename", "counterTask", "Name of child process")
	flag.StringVar(&fOutPutFile, "output", "counts.json", "Output file name")
	flag.Parse()

	// Buil TCP Server
	server, err := intranet.NewServer()
	if err != nil {
		fmt.Println("TCP biding error: ", err)
	} else {
		fmt.Println("listening TCP port on: ", server.GetPort())
	}

	// Run TCP server and set received process
	server.SetRecvCallBack(readRecvMap)

	go server.Listen()

	// Build specificed number of sub proccesses
	_, err = buildSubProccess(fProccessNum, server.GetPort())
	if err != nil {
		fmt.Println("Build sub process error: ", err)
		os.Exit(5)
	}

	go sendRequest(server)

	// Set channel
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	errorCh := server.GetErrorChannel()

	for {
		select {
		case err := <-errorCh:
			fmt.Println("Server TCP error: ", err)
			os.Exit(1)
		case <-sig:
			server.Done()
			saveJsonData()
			return
		}
	}
}

func sendRequest(s intranet.Server) {
	for range time.Tick(1 * time.Second) {
		currentUnixSec := time.Now().Unix()
		pkg := &intranet.CountPackage{
			Time: currentUnixSec,
		}
		buf := bytes.NewBuffer(nil)
		_ = gob.NewEncoder(buf).Encode(&pkg)

		s.Send(buf.Bytes())
	}
}

func buildSubProccess(fProccessNum int, port int) ([]*os.Process, error) {
	// Create sub process
	childPs = make([]*os.Process, fProccessNum)
	for i := 0; i < fProccessNum; i++ {
		paramPort := fmt.Sprintf("--port=%d", port)
		paramThNum := fmt.Sprintf("--t=%d", fThreadNum)

		cmd := exec.Command(fmt.Sprintf("./%s", fProccessName), paramPort, paramThNum)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			NoInheritHandles: false,
		}
		err := cmd.Start()
		if err != nil {
			return nil, fmt.Errorf("run sub proccess error: %s", err.Error())
		}
		childPs = append(childPs, cmd.Process)
	}

	return childPs, nil
}

func readRecvMap(data []byte, errCh chan error) {
	pkg := &intranet.CountPackage{}
	buf := bytes.NewBuffer(data)
	err := gob.NewDecoder(buf).Decode(&pkg)

	if err != nil && err != io.EOF {
		errCh <- fmt.Errorf("parse received data error: %s", err.Error())
		return
	}

	savedData, ok := countMap.Load(pkg.Time)

	if !ok {
		cinfo := CountInfo{1, pkg.Data}
		countMap.Store(pkg.Time, cinfo)
	} else {
		countData := savedData.(CountInfo)
		countData.Count++

		// If reciver has been recived data from all of sub proceess, save data to json string and remove from array
		if countData.Count == fProccessNum {
			jsonData := &JsonData{
				Time:   pkg.Time,
				Counts: make(map[int]int, 10),
			}
			for i, val := range pkg.Data {
				jsonData.Counts[i] = countData.Data[i] + val
			}
			jsonStr, _ := json.Marshal(jsonData)

			// Addend comma
			jsonStr = append(jsonStr, 44)
			totalCountJson = append(totalCountJson, jsonStr...)

			// Delete cache from map
			countMap.Delete(pkg.Time)
		} else {
			for i, val := range pkg.Data {
				countData.Data[i] += val
			}
			countMap.Store(pkg.Time, countData)
		}
	}
}

func saveJsonData() {
	totalCountJson = totalCountJson[:len(totalCountJson)-1]
	totalCountJson = append(totalCountJson, []byte("]")...)

	f, _ := os.OpenFile(fOutPutFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer f.Close()
	f.Write(totalCountJson)
}
