package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Galdoba/gslog"
)

func main() {
	f, _ := os.Create(`./testLog.txt`)
	gslog.SetWriter(f)
	for i := range 999 {
		time.Sleep(time.Millisecond * 70)
		gslog.Info(fmt.Sprintf("message %v", i))
		fmt.Printf("%v\r", i)
	}

}
