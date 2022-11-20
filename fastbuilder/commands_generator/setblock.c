#include "command.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

void SetBlockRequestInternal(GoString *preallocatedStr, int x, int y, int z, const char *blockName, unsigned short data, const char *method) {
	snprintf((char*)preallocatedStr->p,1023,"setblock %d %d %d %s %d %s",x,y,z,blockName,data,method);
	preallocatedStr->n=strlen(preallocatedStr->p);
}

void SetBlockWithBlockStatesRequestInternal(GoString *preallocatedStr, int x, int y, int z, const char *blockName, const char *blockStates, const char *method) {
	snprintf((char*)preallocatedStr->p,1023,"setblock %d %d %d %s %s %s",x,y,z,blockName,blockStates,method);
	preallocatedStr->n=strlen(preallocatedStr->p);
}