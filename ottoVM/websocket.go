package ottoVM

import (
	"github.com/gorilla/websocket"
	"github.com/robertkrimen/otto"
)

func addWebsocket(vm *otto.Otto){
	webscoketConnect:= func(call otto.FunctionCall) (otto.Value) {
		jsAddress:=call.Argument(0)
		address, err := jsAddress.ToString()
		if err != nil {
			return vm.MakeCustomError("webscoketConnect","address is not a valid string")
		}
		jsCallBack:=call.Argument(1)
		if !jsCallBack.IsFunction(){
			return vm.MakeCustomError("webscoketConnect","onMessage is not a valid function")
		}
		conn, _, err := websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			return vm.MakeCustomError("webscoketConnect","cannot connect to address "+address+" error: "+err.Error())
		}
		writeFn:=func(msg string) otto.Value {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				return vm.MakeCustomError("webscoketConnect","cannot connect write to address "+address+" maybe connect is down?")
			}
			return otto.Value{}
		}

		jsWriteFn:= func(call otto.FunctionCall) (otto.Value){
			jsMsg:=call.Argument(0)
			msg, err := jsMsg.ToString()
			if err != nil {
				return vm.MakeCustomError("webscoketWrite","msg is not a valid string")
			}
			return writeFn(msg)
		}

		go func() {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				jsCallBack.Call(otto.UndefinedValue(),vm.MakeCustomError("webscoketRead","fail to read"+err.Error()))
			}else{
				if msgType==websocket.TextMessage{
					jsStr,_:=vm.ToValue(string(data))
					jsCallBack.Call(otto.UndefinedValue(),jsStr)
				}
			}
		}()

		jsWrappedWriteFn,_:=vm.ToValue(jsWriteFn)
		return jsWrappedWriteFn
	}

	vm.Set("_websocketConnectV1",webscoketConnect)

	//if err != nil {
	//	if cq.firstInit {
	//		panic(err)
	//	} else {
	//		log.Println("Go-CQ: CONNECTION ERROR:", err)
	//	}
	//} else {
	//	close(cq.connectLock)
	//	break
	//}
}