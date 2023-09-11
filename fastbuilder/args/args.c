#include <stdio.h>
#include <stdlib.h>
#include <getopt.h>
#include <string.h>
#include <dirent.h>
#include <errno.h>
#ifndef PATH_MAX
#include <limits.h>
#endif
#include <sys/types.h>
#include <stdint.h>
#ifdef WIN32
#ifndef __MINGW32__
#error "This file uses gcc-specific features, please consider switching to mingw gcc."
#endif
#include <windows.h>
#endif

#ifndef FB_VERSION
#define FB_VERSION "(CUSTOMIZED)"
#define FB_COMMIT "???"
#define FB_COMMIT_LONG "???"
#warning "It seems that you're building PhoenixBuilder with plain `go build` command, it is highly recommended to use `make current` instead."
#endif

struct go_string {
	char *buf;
	uint64_t length;
};

#define EMPTY_GOSTRING {"",0}

char args_isDebugMode=0;
char args_disableVersionChecking=0;
struct go_string newAuthServer={
	"https://api.fastbuilder.pro",
	27
};
struct go_string startup_script=EMPTY_GOSTRING;
struct go_string server_code=EMPTY_GOSTRING;
struct go_string server_password=EMPTY_GOSTRING;
struct go_string token_content=EMPTY_GOSTRING;
char args_no_readline=0;
struct go_string custom_gamename=EMPTY_GOSTRING;
char ingame_response=0;
struct go_string listen_address=EMPTY_GOSTRING;

void print_help(const char *self_name) {
	printf("%s [options]\n",self_name);
	printf("\t--debug: Run in debug mode.\n");
	printf("\t-A <url>, --auth-server=<url>: Use the specified authentication server, instead of the default one.\n");
	printf("\t--no-update-checking: Suppress update notification.\n");
	printf("\t-c, --code=<server code>: Specify a server code.\n");
	printf("\t-p, --password=<server password>: Specify the password of the server specified by -c.\n");
	printf("\t-t, --token=<path of FBToken>: Specify the path of FBToken, and quit if the file is unaccessible.\n");
	printf("\t-T, --plain-token=<token>: Specify the token content.\n");
	printf("\t--no-readline: Suppress user input.\n");
	printf("\t-N, --gamename <name>: Specify the game name to use interactive commands (e.g. get), instead of using the server provided one.\n");
	printf("\t--ingame-response: Turn on the feature to listen to commands or give output in game.\n");
	printf("\t--del-userdata: Remove user data and exit.\n");
	printf("\t-L, --listen <address>: Listen and handle WebSocket connection on given address, this will also suppress readline.\n");
	printf("\n");
	printf("\t-h, --help: Show this help context.\n");
	printf("\t-v, --version: Show the version information of this program.\n");
	printf("\t\t--version-plain: Show the version of this program.\n");
}

char *commit_hash() {
	return FB_COMMIT_LONG;
}

void print_version(int detailed) {
	if(!detailed) {
		printf(FB_VERSION "\n");
		return;
	}
	printf("PhoenixBuilder " FB_VERSION "\n");
	printf("COMMIT " FB_COMMIT_LONG "\n");
	printf("\n");
}

void read_token(char *token_path) {
	FILE *file=fopen(token_path,"rb");
	if(!file) {
		fprintf(stderr, "Failed to read token at %s.\n",token_path);
		exit(21);
	}
	fseek(file,0,SEEK_END);
	size_t flen=ftell(file);
	fseek(file,0,SEEK_SET);
	token_content.length=flen;
	token_content.buf=malloc(flen);
	fread(token_content.buf, 1, flen, file);
	fclose(file);
}

void quickmake(struct go_string **target_ptr) {
	size_t length=strlen(optarg);
	char *data=malloc(length);
	memcpy(data, optarg, length);
	*target_ptr=malloc(16);
	(*target_ptr)->buf=data;
	(*target_ptr)->length=length;
}

void quickset(struct go_string *target_ptr) {
	size_t length=strlen(optarg);
	char *data=malloc(length);
	memcpy(data, optarg, length);
	target_ptr->buf=data;
	target_ptr->length=length;
}

#ifdef DT_UNKNOWN

void rmdir_recursive(char *path) {
	char *pathend=path+strlen(path);
	DIR *fbdir=opendir(path);
	if(!fbdir) {
		if(errno==ENOENT) {
			return;
		}
		fprintf(stderr, "Failed to open directory [%s]: %s\n", path, strerror(errno));
		exit(1);
	}
	struct dirent *dir_ent;
	while((dir_ent=readdir(fbdir))!=NULL) {
		if(dir_ent->d_type==DT_UNKNOWN) {
			fprintf(stderr, "Found file with unknown type: %s\n", path);
			exit(2);
		}
		if(dir_ent->d_type!=DT_DIR) {
			sprintf(pathend,"%s",dir_ent->d_name);
			remove(path);
		}else{
			if((dir_ent->d_name[0]=='.'&&dir_ent->d_name[1]==0)||(dir_ent->d_name[0]=='.'&&dir_ent->d_name[1]=='.'&&dir_ent->d_name[2]==0)){
				continue;
			}
			sprintf(pathend,"%s/",dir_ent->d_name);
			rmdir_recursive(path);
			remove(path);
		}
	}
	closedir(fbdir);
	*pathend=0;
}

#else

void go_rmdir_recursive(char *path);
void rmdir_recursive(char *path) {
	go_rmdir_recursive(path);
}

#endif

void config_cleanup() {
	char *home_dir=getenv("HOME");
	if(home_dir==NULL) {
		fprintf(stderr, "Failed to obtain user's home directory, using \".\" instead.\n");
		home_dir=".";
	}
	char *buf=malloc(PATH_MAX);
	sprintf(buf, "%s", home_dir);
	char *concat_start=buf+strlen(buf);
	sprintf(concat_start,"/.config/fastbuilder/");
	rmdir_recursive(buf);
	remove(buf);
	free(buf);
	exit(0);
}

int _parse_args(int argc, char **argv) {
	while(1) {
		static struct option opts[]={
			{"debug", no_argument, 0, 0}, // 0
			{"help", no_argument, 0, 'h'}, // 1
			{"auth-server", required_argument, 0, 'A'}, // 2
			{"no-update-checking", no_argument, 0, 0}, // 3
			{"version", no_argument, 0, 'v'}, // 4
			{"version-plain", no_argument, 0, 0}, // 5
			{"code", required_argument, 0, 'c'}, // 6
			{"password", required_argument, 0, 'p'}, // 7
			{"token", required_argument, 0, 't'}, // 8
			{"plain-token", required_argument, 0, 'T'}, // 9
			{"no-readline", no_argument, 0, 0}, // 10
			{"gamename", required_argument, 0, 'N'}, // 11
			{"ingame-response", no_argument, 0, 0}, // 12
			{"purge-userdata", no_argument, 0, 0}, // 13
			{"listen", required_argument, 0, 'L'}, // 14
			{0, 0, 0, 0}
		};
		int option_index;
		int c=getopt_long(argc,argv,"hA:vc:p:t:T:N:L:", opts, &option_index);
		if(c==-1)
			break;
		switch(c) {
		case 0:
			switch(option_index) {
			case 0:
				args_isDebugMode=1;
				break;
			case 3:
				args_disableVersionChecking=1;
				break;
			case 5:
				print_version(0);
				return 0;
			case 10:
				args_no_readline=1;
				break;
			case 12:
				ingame_response=1;
				break;
			case 13:
				config_cleanup();
				break;
			};
			break;
		case 'h':
			print_help(argv[0]);
			return 0;
		case 'A':
			quickset(&newAuthServer);
			break;
		case 'c':
			quickset(&server_code);
			break;
		case 'p':
			quickset(&server_password);
			break;
		case 't':
			read_token(optarg);
			break;
		case 'T':
			quickset(&token_content);
			break;
		case 'v':
			print_version(1);
			return 0;
		case 'N':
			quickset(&custom_gamename);
			break;
		case 'L':
			args_no_readline=1;
			quickset(&listen_address);
			break;
		default:
			print_help(argv[0]);
			return 1;
		};
	};
	return -1;
}

struct go_string args_var_fbversion_struct={
	FB_VERSION " (" FB_COMMIT ")",
	sizeof(FB_VERSION " (" FB_COMMIT ")")-1
};

struct go_string args_var_fbplainversion_struct={
	FB_VERSION,
	sizeof(FB_VERSION)-1
};

struct go_string args_fb_commit_struct={
	FB_COMMIT,
	sizeof(FB_COMMIT)-1
};

int args_has_specified_server() {
	return server_code.length!=0;
}

int args_specified_token() {
	return token_content.length!=0;
}

#ifndef WIN32
__attribute__((constructor)) static void parse_args(int argc, char **argv) {
	int ec;
	if((ec=_parse_args(argc,argv))!=-1) {
		exit(ec);
	}
	return;
}
#else
__attribute__((constructor)) static void parse_args_win32() {
	int argc;
	char **argv;
	wchar_t **ugly_argv=CommandLineToArgvW(GetCommandLineW(), &argc);
	argv=malloc(sizeof(char*)*argc);
	for(int i=0;i<argc;i++) {
		int len=WideCharToMultiByte(CP_UTF8, 0, ugly_argv[i], -1, NULL, 0, NULL, 0);
		argv[i]=malloc(len);
		WideCharToMultiByte(CP_UTF8, 0, ugly_argv[i], -1, argv[i], len, NULL, 0);
	}
	int ec;
	if((ec=_parse_args(argc,argv))!=-1) {
		exit(ec);
	}
	for(int i=0;i<argc;i++) {
		free(argv[i]);
	}
	free(argv);
	LocalFree(ugly_argv);
	HMODULE winmm_lib=LoadLibraryA("winmm.dll");
	void (*timeBeginPeriod)(int)=(void *)GetProcAddress(winmm_lib, "timeBeginPeriod");
	timeBeginPeriod(1);
	FreeLibrary(winmm_lib);
}
#endif
