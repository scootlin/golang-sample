package main

import (
	"fmt"
	"sncp/internal/intranet"
	"sync"
)

func main() {
	listenPort, err := intranet.LisetnTCP()
	if listenPort != 0 {
		fmt.Println("listening TCP port on: ", listenPort)
	} else {
		fmt.Println("TCP biding error: ", err)
	}
}

func mergerCounter(source, desc *sync.Map) {
	source.Range(func(key, value any) bool {
		currentCount, _ := desc.Load(key.(int))
		currentCount = currentCount.(int) + 1
		return true
	})
}
