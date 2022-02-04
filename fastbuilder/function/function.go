package function

import (
	"strings"
	"fmt"
	"strconv"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/minecraft"
)

type Function struct {
	Name string
	OwnedKeywords []string
	
	FunctionType byte
	SFMinSliceLen uint16
	SFArgumentTypes []byte
	FunctionContent interface{} // Regular/Simple: func(*minecraft.Conn,interface{})
				    // Continue: map[string]*FunctionChainItem
}

type FunctionChainItem struct {
	FunctionType byte
	ArgumentTypes []byte
	Content interface{}
}

const (
	FunctionTypeSimple    = 0 // End of simple chain
	FunctionTypeContinue  = 1 // Simple chain
	FunctionTypeRegular   = 2
)

const (
	SimpleFunctionArgumentString  = 0
	SimpleFunctionArgumentDecider = 1
	SimpleFunctionArgumentInt     = 2
	SimpleFunctionArgumentMessage = 3
	SimpleFunctionArgumentEnum    = 4
)

var FunctionMap = make(map[string]*Function)

func RegisterFunction(function *Function) {
	for _, nm := range function.OwnedKeywords {
		if _, ok := FunctionMap[nm]; !ok {
			FunctionMap[nm]=function
		}
	}
}

type EnumInfo struct {
	WantedValuesDescription string // "discrete, continuous, none"
	Parser func(string)byte
	InvalidValue byte
}

var SimpleFunctionEnums []*EnumInfo

func RegisterEnum(desc string,parser func(string)byte,inv byte) int {
	SimpleFunctionEnums=append(SimpleFunctionEnums,&EnumInfo{WantedValuesDescription:desc,InvalidValue:inv,Parser:parser})
	return len(SimpleFunctionEnums)-1+SimpleFunctionArgumentEnum
}

func Process(conn *minecraft.Conn,msg string) {
	slc:=strings.Split(msg, " ")
	fun, ok := FunctionMap[slc[0]]
	if !ok {
		return
	}
	if fun.FunctionType == FunctionTypeRegular {
		cont, _:=fun.FunctionContent.(func(*minecraft.Conn,string))
		cont(conn, msg)
		return
	}
	if len(slc) < int(fun.SFMinSliceLen) {
		command.Tellraw(conn, fmt.Sprintf("Parser: Simple function %s required at least %d arguments, but got %d.",fun.Name, fun.SFMinSliceLen, len(slc)))
		return
	}
	var arguments []interface{}
	ic:=1
	cc:=&FunctionChainItem {
		FunctionType: fun.FunctionType,
		ArgumentTypes: fun.SFArgumentTypes,
		Content: fun.FunctionContent,
	}
	for {
		if cc.FunctionType == FunctionTypeContinue {
			if len(slc)<=ic {
				rf, _:=cc.Content.(map[string]*FunctionChainItem)
				itm, got := rf[""]
				if !got {
					command.Tellraw(conn, I18n.T(I18n.SimpleParser_Too_few_args))
					return
				}
				cc=itm
				continue
			}
			rfc, _:=cc.Content.(map[string]*FunctionChainItem)
			chainitem, got := rfc[slc[ic]]
			if !got {
				command.Tellraw(conn, I18n.T(I18n.SimpleParser_Invalid_decider))
				return
			}
			cc=chainitem
			ic++
			continue
		}
		if len(cc.ArgumentTypes) > len(slc)-ic {
			command.Tellraw(conn, I18n.T(I18n.SimpleParser_Too_few_args))
			return
		}
		for _, tp := range cc.ArgumentTypes {
			if tp==SimpleFunctionArgumentString {
				arguments=append(arguments,slc[ic])
			}else if tp==SimpleFunctionArgumentDecider {
				command.Tellraw(conn, "Parser: Internal error - argument type [decider] is preserved.")
				fmt.Println("Parser: Internal error - DO NOT REGISTER Decider ARGUMENT!")
				return
			}else if tp==SimpleFunctionArgumentInt {
				parsedInt, err := strconv.Atoi(slc[ic])
				if err != nil {
					command.Tellraw(conn, fmt.Sprintf("%s: %v", I18n.T(I18n.SimpleParser_Int_ParsingFailed), err))
					return
				}
				arguments=append(arguments,parsedInt)
			}else if tp==SimpleFunctionArgumentMessage {
				messageContent:=strings.Join(slc[ic:]," ")
				arguments=append(arguments,messageContent)
				// Arguments after the message argument isn't allowed.
				break
			}else{
				eindex:=int(tp-SimpleFunctionArgumentEnum)
				if eindex>=len(SimpleFunctionEnums) {
					command.Tellraw(conn, "Parser: Internal error, unregistered enum")
					fmt.Printf("Internal error, unregistered enum %d\n",int(tp))
					return
				}
				ei:=SimpleFunctionEnums[eindex]
				itm:=ei.Parser(slc[ic])
				if itm == ei.InvalidValue {
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.SimpleParser_InvEnum),ei.WantedValuesDescription))
					return
				}
				arguments=append(arguments,itm)
			}
			ic++
		}
		cont, _:=cc.Content.(func(*minecraft.Conn,[]interface{}))
		if cont==nil {
			cont,_:=cc.Content.(func(interface{},[]interface{}))
			cont(conn, arguments)
			return
		}
		cont(conn, arguments)
		return
	}
}




