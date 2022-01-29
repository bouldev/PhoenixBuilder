#include "command.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

void replaceItemRequestInternal(GoString *preallocatedStr, int x, int y, int z, unsigned char slot, const char *name, unsigned char count, unsigned short damage) {
	sprintf((char*)preallocatedStr->p,"replaceitem block %d %d %d slot.container %d %s %d %d", x, y, z, slot, name, count, damage);
	preallocatedStr->n=strlen(preallocatedStr->p);
}

