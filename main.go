package main

import (
	"fmt"
	"sort"
)

type Car struct{
	Name string
	Year int
	Factory string
}

type CarSlice []Car//这个存放放Car的切片

func (c CarSlice) Len() int{
	return len(c)
}

func (c CarSlice) Less(i,j int) bool{
	return c[i].Year<c[j].Year//按年份排列
	//return c[i].Name<c[j].Name//按名字排列
}

func (c CarSlice) Swap(i,j int){
	c[i],c[j]=c[j],c[i]
}

func main(){
	var cl CarSlice=[]Car{
		{"collola",2020,"Toyota"},
		{"x7",2018,"BMW"},
		{"CH-V",2018,"Honda"},
		{"XXI",2002,"Dogde"},
		{"LALA93",1998,"Ford"},
	}

	//sort.Sort(cl)//sort是包名，Sort是方法名,我们这里实现的是type Interface interface的接口
	//
	//for _,v:=range cl{
	//	fmt.Println(v)
	//}

	var inter interface{}
	inter=cl

	if value,ok:=inter.(sort.Interface);ok{
			fmt.Println(value,ok)
	}

}