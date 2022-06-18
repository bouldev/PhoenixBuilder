#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void *phoenixbuilder_create();
int phoenixbuilder_execute(void *, char *);
void phoenixbuilder_destroy(void *);
void phoenixbuilder_command_output(void *ref, char *uuid, int succ, char *message, char *param);

void *holdingRef;

void phoenixbuilder_output(char *content) {
	printf("Output>> %s\n", content);
	free(content);
}

void phoenixbuilder_worldchat_output(char *content) {
	printf("WORLDCHAT OUTPUT>> %s\n", content);
	free(content);
}

void phoenixbuilder_send_silent_command(char *command) {
	printf("Silent Command>> %s\n", command);
	free(command);
}

void phoenixbuilder_send_ws_command(char *command, char *uuid) {
	printf("WSCommand>> %s\nWSCommand UUID: %s\n\n",command,uuid);
	phoenixbuilder_command_output(holdingRef, uuid, 1, "", "");
	free(command);
	free(uuid);
}

void phoenixbuilder_send_command(char *command, char *uuid) {
	printf("NormalCommand>> %s\nNormalCommand UUID: %s\n\n",command,uuid);
	phoenixbuilder_command_output(holdingRef, uuid, 1, "", "");
	free(command);
	free(uuid);
}

void phoenixbuilder_send_chat(char *content) {
	printf("CHAT>> %s\n", content);
	free(content);
}

void phoenixbuilder_show_title(char *message) {
	printf("TITLE>> %s\n", message);
	free(message);
}

void *phoenixbuilder_get_ranged_blocks(int a, int b, int c, int d, int e, int f) {
	printf("phoenixbuilder_get_ranged_blocks: Called, returning 0x0 (unsupported)\n");
	return NULL;
}

void phoenixbuilder_get_block() {
	abort();
}

void phoenixbuilder_get_block_data() {
	abort();
}

void phoenixbuilder_get_block_nbt() {
	abort();
}

void phoenixbuilder_update_command_block(int x,int y,int z,unsigned int mode,char *command,char *customName,char *lastOutput,int tickDelay,char executeOnFirstTick,char trackOutput,char conditional,char needsRedstone) {
	printf("phoenixbuilder_update_command_block called\n");
}


void main() {
	printf("TWEAK MODE TEST -->\n");
	printf("Creating PBEnvironment\n");
	void *pbref=phoenixbuilder_create();
	holdingRef=pbref;
	printf("Got reference (not a pointer, no dereference plz): %p\n", pbref);
	char buf[1024];
	char *bufptr=buf;
	printf("Enter command below -->\n");
	while(1) {
		*bufptr=getchar();
		if(*bufptr=='\n') {
			*bufptr=0;
			if(!strcmp(buf, "exit")||!strcmp(buf, "fbexit")) {
				// The original one will try to call conn.Close(),
				// and cause it to crash.
				printf("phoenixbuilder_destroy(%p);\n", pbref);
				phoenixbuilder_destroy(pbref);
				printf("PBEnvironment destroyed\n");
				exit(0);
			}
			printf("Calling phoenixbuilder_execute(%p, \"%s\");\n", pbref, buf);
			bufptr=buf;
			int ret=phoenixbuilder_execute(pbref, buf);
			if(!ret) {
				printf("RETURN: 0 (Command not found)\n");
			}else{
				printf("RETURN: 1 (Found)\n");
			}
			continue;
		}
		bufptr++;
	}
	exit(0);
}