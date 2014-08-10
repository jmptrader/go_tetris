package main

import (
	"fmt"
	"reflect"
)

func init() {
	initFlags()
	initConf()
	initClient()
	initOptions()
}

var (
	t       = reflect.TypeOf(fm{})
	v       = []reflect.Value{reflect.ValueOf(fm{})}
	options string
)

func main() {
	for {
		fmt.Println(options)
		var op int
		_, err := fmt.Scanln(&op)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if op >= t.NumMethod() {
			fmt.Println("the op is not correct")
			continue
		}
		t.Method(op).Func.Call(v)
	}
}

func initOptions() {
	options = "select one option from the following:\n"
	for i := 0; i < t.NumMethod(); i++ {
		options += fmt.Sprintf("%d. %s\n", i, t.Method(i).Name)
	}
}

type fm struct{}

func (fm) GetAll() {
	nts, err := stub.GetAll()
	if err != nil {
		fmt.Println("can not get all: ", err)
		return
	}
	fmt.Println("get all next tournaments: ", nts)
}

func (fm) Add() {
	fmt.Println("please input number_of_candidates, award_to_gold_in_mBTC, award_to_silver_in_mBTC, sponsor...")
	var nc, ag, as int
	var sponsor string
	if _, err := fmt.Scanln(&nc); err != nil {
		fmt.Println("can not scan num of candidate:", err)
		return
	}
	if _, err := fmt.Scanln(&ag); err != nil {
		fmt.Println("can not scan award to gold:", err)
		return
	}
	if _, err := fmt.Scanln(&as); err != nil {
		fmt.Println("can not scan award to silver:", err)
		return
	}
	if _, err := fmt.Scanln(&sponsor); err != nil {
		fmt.Println("can not scan sponsor:", err)
		return
	}
	if err := stub.Add(nc, ag, as, sponsor); err != nil {
		fmt.Println("can not add:", err)
		return
	}
	fmt.Println("successfully add next tournament")
}

func (fm) Delete() {
	if err := stub.Delete(); err != nil {
		fmt.Println("can not delete:", err)
		return
	}
	fmt.Println("successfully delete...")
}
