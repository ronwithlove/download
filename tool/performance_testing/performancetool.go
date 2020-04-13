package main

import (
	"github.com/performance-testing-tool/go-dappley/tool/performance_testing/service"
	"time"
)

const goCount 		= 5//go程数量


func main() {
	dappSdk := service.NewServiceClient()

	for i:=0;i<goCount;i++{
		dappSdk.StartTransactionGoroutine(i+1)
			time.Sleep(5 * time.Second)
	}
	//todo:更新方式需要改一下

	neverQuit := make(chan bool)
	<-neverQuit

}

