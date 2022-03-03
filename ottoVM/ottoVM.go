package ottoVM

import (
	_ "embed"
	"fmt"
	"github.com/robertkrimen/otto"
)

type Runnable interface {
	// start running until an error occour or complete
	// if no error, final result will be returned in a json string
	Run() (string,error)
	RunInRoutine(OnResultCallback func(Result string,err error))

	GetVM() *otto.Otto
	GetName() string
}

type OttoKeeper interface {
	// try to compile script and then insert host(golang) func
	LoadNewScript(script string,name string) Runnable
	// actually is a pack of VM.Set, but each Runnable will be automatically set this
	SetInitFn(func(r Runnable))
}

//go:embed bootstrap.js
var bootStrapJsScript []byte

type RunnableAlpha struct {
	Name string
	VM             *otto.Otto
	Script string
}

func (runnable *RunnableAlpha) Run() (strResult string,err error) {
	var finalVal interface{}
	var jsonResult otto.Value
	Errorf:=func(fmtStr string,a ...interface{}) error {
		return fmt.Errorf("JS-Script(%v): "+fmtStr,runnable.Name,a)
	}
	defer func() {
		r:=recover()
		if r!=nil{
			err=Errorf(fmt.Sprintf("%v",r))
		}
	}()
	finalVal, err = runnable.VM.Run(runnable.Script)
	if err != nil {
		return
	}else{
		err = runnable.VM.Set("finalVal", finalVal)
		if err != nil {
			err= Errorf("cannot set final result (%v)",err)
			return
		}
		jsonResult, err =runnable.VM.Run("JSON.stringify(finalVal)")
		if err != nil {
			err= Errorf("cannot stringify final result (%v)",err)
			return
		}
		strResult, err = jsonResult.ToString()
		if err != nil {
			err= Errorf("cannot get final result (%v)",err)
			return
		}
		return
	}
}

func (Runnable *RunnableAlpha) RunInRoutine(OnResultCallback func(Result string,err error)){
	go func() {
		result, err := Runnable.Run()
		if err != nil {
			OnResultCallback("",fmt.Errorf("RuntimeError %v",err))
		}else{
			OnResultCallback(result,err)
		}
	}()
}

func (runnable *RunnableAlpha) GetVM() *otto.Otto{
	return runnable.VM
}

func (runnable *RunnableAlpha) GetName() string{
	return runnable.Name
}

type OttoKeeperAlpha struct {
	initFn func(r Runnable)
}

func (oa *OttoKeeperAlpha)LoadNewScript(script string,name string) Runnable {
	script=string(bootStrapJsScript)+"\n"+script
	vm:=otto.New()
	ra:=&RunnableAlpha{Name: name,VM: vm,Script:script}
	oa.initFn(ra)
	return ra
}

func (oa *OttoKeeperAlpha) SetInitFn(initFn func(r Runnable)){
	oa.initFn=initFn
}