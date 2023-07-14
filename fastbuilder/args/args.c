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
char args_disableVersionCheck=0;
struct go_string newAuthServer={
	"https://api.fastbuilder.pro",
	27
};
struct go_string startup_script=EMPTY_GOSTRING;
struct go_string server_code=EMPTY_GOSTRING;
struct go_string server_password=EMPTY_GOSTRING;
struct go_string token_content=EMPTY_GOSTRING;
struct go_string externalListenAddr=EMPTY_GOSTRING;
struct go_string capture_output_file=EMPTY_GOSTRING;
char args_no_readline=0;
struct go_string pack_scripts=EMPTY_GOSTRING;
struct go_string pack_scripts_out=EMPTY_GOSTRING;
char enable_omega_system=0;
struct go_string custom_gamename=EMPTY_GOSTRING;
char ingame_response=0;

extern void custom_script_engine_const(const char *key, const char *val);
extern void do_suppress_se_const(const char *key);

// TODO: Localizations via Gettext/Glibc intl
void print_help(const char *self_name) {
	printf("%s [options]\n",self_name);
	printf("\t--debug: Run in debug mode.\n");
	printf("\t-A <url>, --auth-server=<url>: Use the specified authentication server, instead of the default one.\n");
	printf("\t--no-update-check: Suppress update notifications.\n");
	printf("\t--force-pyrpc: Enable the PyRpcPacket interaction, client will be kicked automatically by netease's rental server.\n");
#ifdef WITH_V8
	printf("\t-S, --script=<*.js>: run a .js script at start\n");
	printf("\t--script-engine-const key=value: Define a const value for script engine's \"consts\" const. Can be used to replace the default value. Specify multiple items by using this argument for multiple times.\n");
	printf("\t--script-engine-suppress-const <key>: Undefine a const value for script engine's \"consts\" const. Specify multiple items by using this argument for multiple times.\n");
#endif
	printf("\t-c, --code=<server code>: Specify a server code.\n");
	printf("\t-p, --password=<server password>: Specify the password of the server specified by -c.\n");
	printf("\t-t, --token=<path of FBToken>: Specify the path of FBToken, and quit if the file is unaccessible.\n");
	printf("\t-T, --plain-token=<token>: Specify the token content.\n");
	printf("\t-E, --listen-external: Listen on the specified address and wait for external controlling connection.\n\t\tExample: -E 0.0.0.0:5768 - listen on port 5768 and accept connections from anywhere,\n\t\t\t-E 127.0.0.1:5769 - listen on port 5769 and accept connections from localhost only.\n");
	printf("\t--capture=<*.bin>: Capture minecraft packet and dump to target file\n");
	printf("\t--no-readline: Suppress user input.\n");
	printf("\t--pack-scripts <manifest path>: Create a script package.\n");
	printf("\t--pack-scripts-to <path>: Specify the path for the output script package.\n");
	printf("\t-N, --gamename <name>: Specify the game name to use interactive commands (e.g. get), instead of using the server provided one.\n");
	printf("\t--ingame-response: Turn on the feature to listen to commands or give output in game.\n");
	printf("\t--del-userdata: Remove user data and exit.\n");
	printf("\n");
	printf("\t-O, --omega_system: Enable Omega System.\n");
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
#ifdef FBGUI_VERSION
	printf("With GUI " FBGUI_VERSION "\n");
#endif
#ifdef WITH_V8
	printf("With V8 linked.\n");
#endif
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

void quickcopy(char **target_ptr) {
	size_t length=strlen(optarg)+1;
	*target_ptr=malloc(length);
	memcpy(*target_ptr, optarg, length);
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
			{"auth-server", required_argument, 0, 'A'}, //2
			{"no-update-check", no_argument, 0, 0}, //3
			{"no-world-chat", no_argument, 0, 'M'}, //4
			{"force-pyrpc", no_argument, 0, 0}, //5
			{"no-nbt", no_argument, 0, 0}, //6
			{"script", required_argument, 0, 'S'}, //7
			{"version", no_argument, 0, 'v'}, //8
			{"version-plain", no_argument, 0, 0}, //9
			{"code", required_argument, 0, 'c'}, //10
			{"password", required_argument, 0, 'p'}, //11
			{"token", required_argument, 0, 't'}, //12
			{"plain-token", required_argument, 0, 'T'}, //13
			{"script-engine-const", required_argument, 0, 0}, //14
			{"script-engine-suppress-const", required_argument, 0, 0}, //15
			{"listen-external", required_argument, 0, 'E'}, // 16
			{"no-readline", no_argument, 0, 0}, //17
			{"pack-scripts", required_argument, 0, 0}, //18
			{"pack-scripts-to", required_argument, 0, 0}, //19
			{"capture", required_argument, 0, 0}, // 20
			{"omega_system", no_argument, 0, 'O'}, // 21
			{"gamename", required_argument, 0, 'N'}, // 22
			{"ingame-response", no_argument, 0, 0}, // 23
			{"del-userdata", no_argument, 0, 0}, // 24
			{0, 0, 0, 0}
		};
		int option_index;
		int c=getopt_long(argc,argv,"hA:MvS:c:p:t:T:ON:", opts, &option_index);
		if(c==-1)
			break;
		switch(c) {
		case 0:
			switch(option_index) {
			case 0:
				args_isDebugMode=1;
				break;
			case 3:
				args_disableVersionCheck=1;
				break;
			case 5:
				fprintf(stderr, "--force-pyrpc not available\n");
				return 10;
				break;
			case 6:
				fprintf(stderr, "--no-nbt option is no longer available.\n");
				return 10;
				break;
			case 9:
				print_version(0);
				return 0;
			case 14:
#ifndef WITH_V8
				fprintf(stderr,"--script-engine-const argument isn't available: Non-v8-linked version.\n");
				return 10;
#endif
				{
					int break_switch_14=0;
					for(char *ptr=optarg;*ptr!=0;ptr++) {
						if(*ptr=='=') {
							*ptr=0;
							ptr++;
							custom_script_engine_const(optarg, ptr);
							break_switch_14=1;
							break;
						}
					}
					if(break_switch_14)break;
					fprintf(stderr, "--script-engine-const: Format: key=val\n");
					print_help(argv[0]);
					return 1;
				}
			case 15:
#ifndef WITH_V8
				fprintf(stderr,"--script-engine-suppress-const argument isn't available: Non-v8-linked version.\n");
				return 10;
#endif
				do_suppress_se_const(optarg);
				break;
			case 17:
				args_no_readline=1;
				break;
			case 18:
				quickset(&pack_scripts);
				break;
			case 19:
				quickset(&pack_scripts_out);
				break;
			case 20:
				quickset(&capture_output_file);
				break;
			case 23:
				ingame_response=1;
				break;
			case 24:
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
		case 'S':
#ifndef WITH_V8
			fprintf(stderr,"-S, --script option isn't available: No V8 linked for this version.\n");
			return 10;
#endif
			quickset(&startup_script);
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
		case 'E':
			quickset(&externalListenAddr);
			break;
		case 'v':
			print_version(1);
			return 0;
		case 'O':
			enable_omega_system=1;
			break;
		case 'N':
			quickset(&custom_gamename);
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

/*
// Go uses a different ABI than C, which would use BX for the 2nd return,
// and we couldn't do that w/o asm.
struct go_string *args_func_authServer() {
	if(!newAuthServer) {
		static struct go_string original_auth_server={
			"wss://api.fastbuilder.pro:2053/",
			31
		};
		return &original_auth_server;
	}
	return newAuthServer;
}*/

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
