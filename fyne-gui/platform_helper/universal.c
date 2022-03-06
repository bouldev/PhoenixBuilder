#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

void initStdoutRedirector(char *rootPath) {
	char pathBuffer[1024]={0};
	sprintf(pathBuffer,"%s/REDIRECT_STDOUT",rootPath);
	if(access(pathBuffer, F_OK)==0) {
		sprintf(pathBuffer,"%s/stdout.log",rootPath);
		freopen(pathBuffer, "wb", stdout);
		sprintf(pathBuffer,"%s/stderr.log",rootPath);
		freopen(pathBuffer, "wb", stderr);
		printf("[StdoutRedirector] Redirected stdout & stderr to filesystem.\n");
	}
	printf("[StdoutRedirector] App's root path is located at %s\n", rootPath);
	free(rootPath);
}