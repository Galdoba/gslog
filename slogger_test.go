package gslog

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func Test_logger_start(t *testing.T) {

	f, _ := os.OpenFile(`/home/galdoba/go/src/github.com/Galdoba/gslog/testing/testLog.txt`, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	cfg := DefaultConfiguration()
	cfg.Handlers["debug"] = NewHandler(f).WithMinimumLevel(DEBUG).WithMaximumLevel(DEBUG)
	debugLogger, err := New(cfg)
	fmt.Println(err)
	// SetWriter(f)
	fmt.Println("start:")
	time.Sleep(time.Second)
	Info("Ping")
	time.Sleep(time.Second)
	fmt.Println("")
	if err := debugLogger.processMessage(mustNewMessage("Ping 2", "answer", 42).WithLevel(DEBUG)); err != nil {
		fmt.Println("error", err)
	}

	fmt.Println("")
}
