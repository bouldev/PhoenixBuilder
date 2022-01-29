#include "command.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

void SetBlockRequestInternal(GoString *preallocatedStr, int x, int y, int z, const char *blockName, unsigned short data, const char *method) {
	sprintf((char*)preallocatedStr->p,"setblock %d %d %d %s %d %s",x,y,z,blockName,data,method);
	preallocatedStr->n=strlen(preallocatedStr->p);
}

