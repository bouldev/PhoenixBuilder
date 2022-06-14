package commands_generator

/*
void *allocateRequestString();
void freeRequestString(void*);
*/
import "C"
import "unsafe"

var AdditionalChatCb func(string) = func(_ string) {}
var AdditionalTitleCb func(string) = func(_ string) {}

func AllocateRequestString() *string {
	return (*string)(C.allocateRequestString())
}

func FreeRequestString(str string) {
	C.freeRequestString(unsafe.Pointer(&str))
}

func FreeRequestStringPtr(str *string) {
	C.freeRequestString(unsafe.Pointer(str))
}
