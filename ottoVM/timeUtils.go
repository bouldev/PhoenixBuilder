package ottoVM

import (
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
)

type _timer struct {
	timer    *time.Timer
	duration time.Duration
	interval bool
	call     otto.FunctionCall
}

// copied from https://github.com/robertkrimen/natto/blob/master/natto.go
func addTimeOut(vm *otto.Otto) error {
	registry := map[*_timer]*_timer{}
	ready := make(chan *_timer)

	newTimer := func(call otto.FunctionCall, interval bool) (*_timer, otto.Value) {
		delay, _ := call.Argument(1).ToInteger()
		if 0 >= delay {
			delay = 1
		}

		timer := &_timer{
			duration: time.Duration(delay) * time.Millisecond,
			call:     call,
			interval: interval,
		}
		registry[timer] = timer

		timer.timer = time.AfterFunc(timer.duration, func() {
			ready <- timer
		})

		value, err := call.Otto.ToValue(timer)
		if err != nil {
			panic(err)
		}

		return timer, value
	}

	setTimeout := func(call otto.FunctionCall) otto.Value {
		_, value := newTimer(call, false)
		return value
	}
	vm.Set("FB_setTimeout", setTimeout)

	setInterval := func(call otto.FunctionCall) otto.Value {
		_, value := newTimer(call, true)
		return value
	}
	vm.Set("FB_setInterval", setInterval)

	clearTimeout := func(call otto.FunctionCall) otto.Value {
		timer, _ := call.Argument(0).Export()
		if timer, ok := timer.(*_timer); ok {
			timer.timer.Stop()
			delete(registry, timer)
		}
		return otto.UndefinedValue()
	}
	vm.Set("FB_clearTimeout", clearTimeout)
	vm.Set("FB_clearInterval", clearTimeout)

	go func() {
		for {
			select {
			case timer := <-ready:
				var arguments []interface{}
				if len(timer.call.ArgumentList) > 2 {
					tmp := timer.call.ArgumentList[2:]
					arguments = make([]interface{}, 2+len(tmp))
					for i, value := range tmp {
						arguments[i+2] = value
					}
				} else {
					arguments = make([]interface{}, 1)
				}
				arguments[0] = timer.call.ArgumentList[0]
				_, err := vm.Call(`Function.call.call`, nil, arguments...)
				if err != nil {
					for _, timer := range registry {
						timer.timer.Stop()
						delete(registry, timer)
						fmt.Printf("VM: timer cb error %v", err)
					}
				}
				if timer.interval {
					timer.timer.Reset(timer.duration)
				} else {
					delete(registry, timer)
				}
			default:
				// Escape valve!
				// If this isn't here, we deadlock...
			}
		}
		fmt.Println("Timer Quit!!!")
	}()
	return nil
}
