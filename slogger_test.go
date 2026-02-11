package gslog

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func Test_logger_start(t *testing.T) {

	// l, err := New(DefaultConfiguration())
	// fmt.Println(err)
	f, _ := os.Create(`/home/galdoba/go/src/github.com/Galdoba/gslog/testing/testLog.txt`)
	SetWriter(f)
	fmt.Println("start:")
	time.Sleep(time.Second)
	Fatal("Ouch!")
	time.Sleep(time.Second)
	fmt.Println("")

	fmt.Println("")
}
