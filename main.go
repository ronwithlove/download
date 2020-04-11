package main

import (
	"fmt"
	"time"
)



func main(){
	ticker:=time.NewTicker(time.Second*100)//定时2秒
	defer ticker.Stop()


	go func(){
		select {
		//当有查到新区块的时候执行
		case <-ticker.C:
			fmt.Println("hello world!")
		}
	}()


	neverQuit:= make(chan bool)
	<-neverQuit
}