package ottoVM

import (
	"fknsrs.biz/p/ottoext/fetch"
	"fknsrs.biz/p/ottoext/loop"
	"fknsrs.biz/p/ottoext/loop/looptask"
	"fknsrs.biz/p/ottoext/process"
	"fknsrs.biz/p/ottoext/promise"
	"flag"
	"fmt"
	"github.com/robertkrimen/otto"
)

// copied from https://github.com/deoxxa/ottoext
func addExt(vm *otto.Otto) {

	l := loop.New(vm)
	blockingTask := looptask.NewEvalTask("")
	//if err := timers.Define(vm, l); err != nil {
	//	fmt.Printf("cannot add timer to VM %v\n",err.Error())
	//}
	if err := promise.Define(vm, l); err != nil {
		fmt.Printf("cannot add promise to VM %v\n",err.Error())
	}
	if err := fetch.Define(vm, l); err != nil {
		fmt.Printf("cannot add fetch to VM %v\n",err.Error())
	}
	if err := process.Define(vm, flag.Args()); err != nil {
		fmt.Printf("cannot add process to VM %v\n",err.Error())
	}
	l.Add(blockingTask)

	go func() {
		if err := l.Run(); err != nil {
			fmt.Printf("An error occours in VM Event loop %v\n",err.Error())
		}
		fmt.Printf("VM Event loop Quit!!!")
	}()
}