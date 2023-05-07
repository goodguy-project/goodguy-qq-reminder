package main

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

var (
	location, _ = time.LoadLocation("Asia/Shanghai")
)

func cronjob() {
	c := cron.New(cron.WithSeconds(), cron.WithLocation(location))
	_, err := c.AddFunc("@every 40m", Gao)
	if err != nil {
		panic(err)
	}
	c.Run()
}

func main() {
	fmt.Println("正常启动")
	go cronjob()
	select {}
}
