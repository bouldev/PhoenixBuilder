package command

/*
void *allocateRequestString();
void freeRequestString(void*);
*/
import "C"
import "unsafe"

func AllocateRequestString() *string {
	return (*string)(C.allocateRequestString())
}

func FreeRequestString(str string) {
	C.freeRequestString(unsafe.Pointer(&str))
}

func FreeRequestStringPtr(str *string) {
	C.freeRequestString(unsafe.Pointer(str))
}