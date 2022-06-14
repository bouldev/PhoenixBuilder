#include "command.h"
#include <stdint.h>
#include <string.h>
#include <stdlib.h>

char *resolveGoStringPtr(GoString *str) {
	static char ret[1024*5]={0};
	// use static so it'll be a pre-allocated space and the content will be
	// replaced every time calling this function.
	memcpy(ret,str->p,str->n);
	ret[str->n]=0;
	return ret;
}

GoString *allocateRequestString() {
	GoString *str=malloc(sizeof(GoString)+sizeof(uint64_t));
	str->n=0;
	str->p=malloc(1024);
	*(void**)((void*)str+sizeof(GoString))=str;
	return str;
}

void freeRequestString(GoString *str) {
	free((char *)str->p);
	free(*(void**)((void*)str+sizeof(GoString)));
}