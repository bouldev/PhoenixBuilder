package function

import (
	"strings"
	"fmt"
	"strconv"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/i18n"
)

type Function struct {
	Name string
	OwnedKeywords []string
	
	FunctionType byte
	SFMinSliceLen uint16
	SFArgumentTypes []byte
	FunctionContent interface{} // Regular/Simple: func(*environment.PBEnvironment,interface{})
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

type EnumInfo struct {
	WantedValuesDescription string // "discrete, continuous, none"
	Parser func(string)byte
	InvalidValue byte
}

type FunctionHolder struct {
	env *environment.PBEnvironment
	FunctionMap map[string]*Function
	SimpleFunctionEnums []*EnumInfo
}

func NewFunctionHolder(env *environment.PBEnvironment) *FunctionHolder {
	return &FunctionHolder {
		env: env,
		FunctionMap: map[string]*Function {},
		SimpleFunctionEnums: []*EnumInfo {},
	}
}

func (function_holder *FunctionHolder) RegisterFunction(function *Function) {
	for _, nm := range function.OwnedKeywords {
		if _, ok := function_holder.FunctionMap[nm]; !ok {
			function_holder.FunctionMap[nm]=function
		}
	}
}

func (function_holder *FunctionHolder) RegisterEnum(desc string,parser func(string)byte,inv byte) int {
	function_holder.SimpleFunctionEnums=append(function_holder.SimpleFunctionEnums,&EnumInfo{WantedValuesDescription:desc,InvalidValue:inv,Parser:parser})
	return len(function_holder.SimpleFunctionEnums)-1+SimpleFunctionArgumentEnum
}

func (function_holder *FunctionHolder) Process(msg string) {
	cmdsender:=function_holder.env.CommandSender
	slc:=strings.Split(msg, " ")
	fun, ok := function_holder.FunctionMap[slc[0]]
	if !ok {
		return
	}
	if fun.FunctionType == FunctionTypeRegular {
		cont, _:=fun.FunctionContent.(func(*environment.PBEnvironment,string))
		cont(function_holder.env, msg)
		return
	}
	if len(slc) < int(fun.SFMinSliceLen) {
		cmdsender.Tellraw(fmt.Sprintf("Parser: Simple function %s required at least %d arguments, but got %d.",fun.Name, fun.SFMinSliceLen, len(slc)))
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
					cmdsender.Tellraw(I18n.T(I18n.SimpleParser_Too_few_args))
					return
				}
				cc=itm
				continue
			}
			rfc, _:=cc.Content.(map[string]*FunctionChainItem)
			chainitem, got := rfc[slc[ic]]
			if !got {
				cmdsender.Tellraw(I18n.T(I18n.SimpleParser_Invalid_decider))
				return
			}
			cc=chainitem
			ic++
			continue
		}
		if len(cc.ArgumentTypes) > len(slc)-ic {
			cmdsender.Tellraw(I18n.T(I18n.SimpleParser_Too_few_args))
			return
		}
		for _, tp := range cc.ArgumentTypes {
			if tp==SimpleFunctionArgumentString {
				arguments=append(arguments,slc[ic])
			}else if tp==SimpleFunctionArgumentDecider {
				cmdsender.Tellraw("Parser: Internal error - argument type [decider] is preserved.")
				fmt.Println("Parser: Internal error - DO NOT REGISTER Decider ARGUMENT!")
				return
			}else if tp==SimpleFunctionArgumentInt {
				parsedInt, err := strconv.Atoi(slc[ic])
				if err != nil {
					cmdsender.Tellraw(fmt.Sprintf("%s: %v", I18n.T(I18n.SimpleParser_Int_ParsingFailed), err))
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
				if eindex>=len(function_holder.SimpleFunctionEnums) {
					cmdsender.Tellraw("Parser: Internal error, unregistered enum")
					fmt.Printf("Internal error, unregistered enum %d\n",int(tp))
					return
				}
				ei:=function_holder.SimpleFunctionEnums[eindex]
				itm:=ei.Parser(slc[ic])
				if itm == ei.InvalidValue {
					cmdsender.Tellraw(fmt.Sprintf(I18n.T(I18n.SimpleParser_InvEnum),ei.WantedValuesDescription))
					return
				}
				arguments=append(arguments,itm)
			}
			ic++
		}
		cont, _:=cc.Content.(func(*environment.PBEnvironment,[]interface{}))
		if cont==nil {
			cont,_:=cc.Content.(func(interface{},[]interface{}))
			if(cont==nil) {
				fmt.Printf("Internal error: invalid type for function\n")
				return
			}
			cont(function_holder.env, arguments)
			return
		}
		cont(function_holder.env, arguments)
		return
	}
}




