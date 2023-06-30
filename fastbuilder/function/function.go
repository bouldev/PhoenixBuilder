package function

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"strconv"
	"strings"
)

type Function struct {
	Name          string
	OwnedKeywords []string

	FunctionType    byte
	SFMinSliceLen   uint16
	SFArgumentTypes []byte
	FunctionContent interface{} // Regular/Simple: func(*environment.PBEnvironment,interface{})
	// Continue: map[string]*FunctionChainItem
}

type FunctionChainItem struct {
	FunctionType  byte
	ArgumentTypes []byte
	Content       interface{}
}

const (
	FunctionTypeSimple   = 0 // End of simple chain
	FunctionTypeContinue = 1 // Simple chain
	FunctionTypeRegular  = 2
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
	Parser                  func(string) byte
	InvalidValue            byte
}

type FunctionHolder struct {
	env                 *environment.PBEnvironment
	FunctionMap         map[string]*Function
	SimpleFunctionEnums []*EnumInfo
}

func NewFunctionHolder(env *environment.PBEnvironment) *FunctionHolder {
	return &FunctionHolder{
		env:                 env,
		FunctionMap:         map[string]*Function{},
		SimpleFunctionEnums: []*EnumInfo{},
	}
}

func (function_holder *FunctionHolder) RegisterFunction(function *Function) {
	for _, nm := range function.OwnedKeywords {
		if _, ok := function_holder.FunctionMap[nm]; !ok {
			function_holder.FunctionMap[nm] = function
		}
	}
}

func (function_holder *FunctionHolder) RegisterEnum(desc string, parser func(string) byte, inv byte) int {
	function_holder.SimpleFunctionEnums = append(function_holder.SimpleFunctionEnums, &EnumInfo{WantedValuesDescription: desc, InvalidValue: inv, Parser: parser})
	return len(function_holder.SimpleFunctionEnums) - 1 + SimpleFunctionArgumentEnum
}

// return: Found bool
func (function_holder *FunctionHolder) Process(msg string) bool {
	slc := strings.Split(msg, " ")
	fun, ok := function_holder.FunctionMap[slc[0]]
	if !ok {
		return false
	}
	if fun.FunctionType == FunctionTypeRegular {
		cont, _ := fun.FunctionContent.(func(*environment.PBEnvironment, string))
		cont(function_holder.env, msg)
		return true
	}
	if len(slc) < int(fun.SFMinSliceLen) {
		function_holder.env.GameInterface.Output(fmt.Sprintf("Parser: Simple function %s required at least %d arguments, but got %d.", fun.Name, fun.SFMinSliceLen, len(slc)))
		return true
	}
	var arguments []interface{}
	ic := 1
	cc := &FunctionChainItem{
		FunctionType:  fun.FunctionType,
		ArgumentTypes: fun.SFArgumentTypes,
		Content:       fun.FunctionContent,
	}
	for {
		if cc.FunctionType == FunctionTypeContinue {
			if len(slc) <= ic {
				rf, _ := cc.Content.(map[string]*FunctionChainItem)
				itm, got := rf[""]
				if !got {
					function_holder.env.GameInterface.Output(I18n.T(I18n.SimpleParser_Too_few_args))
					return true
				}
				cc = itm
				continue
			}
			rfc, _ := cc.Content.(map[string]*FunctionChainItem)
			chainitem, got := rfc[slc[ic]]
			if !got {
				function_holder.env.GameInterface.Output(I18n.T(I18n.SimpleParser_Invalid_decider))
				return true
			}
			cc = chainitem
			ic++
			continue
		}
		if len(cc.ArgumentTypes) > len(slc)-ic {
			function_holder.env.GameInterface.Output(I18n.T(I18n.SimpleParser_Too_few_args))
			return true
		}
		for _, tp := range cc.ArgumentTypes {
			if tp == SimpleFunctionArgumentString {
				arguments = append(arguments, slc[ic])
			} else if tp == SimpleFunctionArgumentDecider {
				function_holder.env.GameInterface.Output("Parser: Internal error - argument type [decider] is preserved.")
				fmt.Println("Parser: Internal error - DO NOT REGISTER Decider ARGUMENT!")
				return true
			} else if tp == SimpleFunctionArgumentInt {
				parsedInt, err := strconv.Atoi(slc[ic])
				if err != nil {
					function_holder.env.GameInterface.Output(fmt.Sprintf("%s: %v", I18n.T(I18n.SimpleParser_Int_ParsingFailed), err))
					return true
				}
				arguments = append(arguments, parsedInt)
			} else if tp == SimpleFunctionArgumentMessage {
				messageContent := strings.Join(slc[ic:], " ")
				arguments = append(arguments, messageContent)
				// Arguments after the message argument isn't allowed.
				break
			} else {
				eindex := int(tp - SimpleFunctionArgumentEnum)
				if eindex >= len(function_holder.SimpleFunctionEnums) {
					function_holder.env.GameInterface.Output("Parser: Internal error, unregistered enum")
					fmt.Printf("Internal error, unregistered enum %d\n", int(tp))
					return true
				}
				ei := function_holder.SimpleFunctionEnums[eindex]
				itm := ei.Parser(slc[ic])
				if itm == ei.InvalidValue {
					function_holder.env.GameInterface.Output(fmt.Sprintf(I18n.T(I18n.SimpleParser_InvEnum), ei.WantedValuesDescription))
					return true
				}
				arguments = append(arguments, itm)
			}
			ic++
		}
		cont, _ := cc.Content.(func(*environment.PBEnvironment, []interface{}))
		if cont == nil {
			cont, _ := cc.Content.(func(interface{}, []interface{}))
			if cont == nil {
				fmt.Printf("Internal error: invalid type for function\n")
				return true
			}
			cont(function_holder.env, arguments)
			return true
		}
		cont(function_holder.env, arguments)
		return true
	}
}
