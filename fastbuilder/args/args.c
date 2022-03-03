#include <stdio.h>
#include <stdlib.h>
#include <getopt.h>
#include <string.h>

// Decided to use --wrap described by GNU's ld(1) at first, but it seems that darwin's 
// ld didn't implement it, so uses objcopy in Makefile instead.

char args_isDebugMode=0;
char args_disableHashCheck=0;
char replaced_auth_server=0;
char *newAuthServer;
char args_muteWorldChat=0;
char args_noPyRpc=0;
char args_noNBT=0;
char use_startup_script=0;
char *startup_script;

void print_help(const char *self_name) {
	printf("%s [options]\n",self_name);
	printf("\t--debug: Run in debug mode.\n");
	printf("\t-A <url>, --auth-server=<url>: Use the specified authentication server, instead of the default one.\n");
	printf("\t--no-hash-check: Disable the hash check.\n");
	printf("\t-M, --no-world-chat: Ignore world chat on client side.\n");
	printf("\t--no-pyrpc: Disable the PyRpcPacket interaction, the client's commands will be prevented from execution by netease's rental server.\n");
	printf("\t--no-nbt: Disable NBT Construction feature.\n");
	printf("\t--script=<*.js>: run a .js script at start");
	printf("\n");
	printf("\t-h, --help: Show this help context.\n");
	printf("\t-v, --version: Show the version information of this program.\n");
	printf("\t\t--version-plain: Show the version of this program.\n");
}

char *get_fb_version() {
	return FB_VERSION " (" FB_COMMIT ")";
}

char *commit_hash() {
	return FB_COMMIT_LONG;
}

void print_version(int detailed) {
	if(!detailed) {
		printf(FB_VERSION "\n");
		return;
	}
	printf("FastBuilder " FB_VERSION "\n");
	printf("COMMIT " FB_COMMIT_LONG "\n");
	printf("Copyright (C) 2022 Bouldev\n");
	printf("\n");
}

int _parse_args(int argc, char **argv) {
	while(1) {
		static struct option opts[]={
			{"debug", no_argument, 0, 0}, // 0
			{"help", no_argument, 0, 'h'}, // 1
			{"auth-server", required_argument, 0, 'A'}, //2
			{"no-hash-check", no_argument, 0, 0}, //3
			{"no-world-chat", no_argument, 0, 'M'}, //4
			{"no-pyrpc", no_argument, 0, 0}, //5
			{"no-nbt", no_argument, 0, 0}, //6
			{"script", required_argument, 0, 'S'}, //7
			{"version", no_argument, 0, 'v'}, //8
			{"version-plain", no_argument, 0, 0}, //9
			{0, 0, 0, 0}
		};
		int option_index;
		int c=getopt_long(argc,argv,"hA:Mv", opts, &option_index);
		if(c==-1)
			break;
		size_t loo=strlen(optarg);
		switch(c) {
		case 0:
			switch(option_index) {
			case 0:
				args_isDebugMode=1;
				break;
			case 3:
				args_disableHashCheck=1;
				break;
			case 5:
				args_noPyRpc=1;
				break;
			case 6:
				args_noNBT=1;
				break;
			case 9:
				print_version(0);
				return 0;
			};
			break;
		case 'h':
			print_help(argv[0]);
			return 0;
		case 'A':
			replaced_auth_server=1;
			newAuthServer=malloc(loo+1);
			memcpy(newAuthServer,optarg,loo+1);
			break;
		case 'M':
			args_muteWorldChat=1;
			break;
		case 'S':
		    use_startup_script=1;
            startup_script=malloc(loo+1);
            memcpy(startup_script,optarg,loo+1);
            break;
		case 'v':
			print_version(1);
			return 0;
		default:
			print_help(argv[0]);
			return 1;
		};
	};
	return -1;
}

void parse_args(int argc, char **argv) {
	int ec;
	if((ec=_parse_args(argc,argv))!=-1) {
		exit(ec);
	}
	return;
}
