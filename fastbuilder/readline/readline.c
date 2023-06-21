// +build !windows,!android android,!arm
// +build !no_readline

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <stdint.h>

#include <readline/readline.h>
#include <readline/history.h>

extern void teardown_self();
extern char **GetFunctionList();
char **fb_readline_completion(const char *text, int start, int end);
char *fb_command_generator(const char *text, int state);
static char *fb_args_generator(const char *text, int state);
static char *fb_delay_first_arg_generator(const char *text, int state);
static char *fb_set_or_get_generator(const char *text, int state);
static char *fb_delay_mode_generator(const char *text, int state);
static char *fb_task_root_args_generator(const char *text, int state);
static char *fb_task_type_generator(const char *text, int state);
static char *fb_bool_value_generator(const char *text, int state);
static char *fb_facing_enum_generator(const char *text, int state);
static char *fb_shape_enum_generator(const char *text, int state);
static char *fb_get_args_generator(const char *text, int state);

static char **tmpFunctionList;

char **strengthenStringArray(const char **source, int entries) {
	char **target=malloc(sizeof(void *)*(entries+1));
	target[entries]=NULL;
	memcpy(target,source,sizeof(void *)*entries);
	return target;
}

static int should_show_fbexit_notice=0;

char *doReadline() {
	char *line=readline("> ");
	if(line==NULL) {
		line=malloc(5);
		line[0]='e';
		line[1]='x';
		line[2]='i';
		line[3]='t';
		line[4]=0;
	}
	add_history(line);
	should_show_fbexit_notice=0;
	return line;
}

void do_sigint_interrupt() {
	printf("^C\n");
	if(rl_end==0) {
		if(should_show_fbexit_notice==1) {
			printf("(To exit, press Ctrl+C again or Ctrl+D or type exit)\n");
		}
		if(should_show_fbexit_notice!=2) {
			should_show_fbexit_notice++;
		}else{
			teardown_self();
			return;
		}
	}else{
		should_show_fbexit_notice=0;
	}
	rl_on_new_line();
	rl_replace_line("",0);
	rl_redisplay();
	return;
}

void do_interrupt() {
	rl_free_line_state ();
	rl_cleanup_after_signal ();
	RL_UNSETSTATE(RL_STATE_ISEARCH|RL_STATE_NSEARCH|RL_STATE_VIMOTION|RL_STATE_NUMERICARG|RL_STATE_MULTIKEY);
	rl_line_buffer[rl_point = rl_end = rl_mark = 0] = 0;
	rl_callback_handler_remove();
	printf("\n");
}

void init_readline() {
	stifle_history(256);
	rl_readline_name="PhoenixBuilder";
	rl_attempted_completion_function=fb_readline_completion;
	rl_catch_signals=0;
	rl_catch_sigwinch=0;
}

void **readline_to_args() {
	// TODO: Escape (\ ) (" ")
	int argc=0;
	char *copied_rl_buffer=malloc(rl_point);
	memcpy(copied_rl_buffer,rl_line_buffer,rl_point);
	for(int i=0;i<rl_point;i++) {
		if(copied_rl_buffer[i]==' ') {
			copied_rl_buffer[i]=0;
			argc++;
		}
	}
	//argc--;
	char **argv=malloc(sizeof(void *)*(argc+1));
	argv[argc]=NULL;
	int current_begin=0;
	int argi=0;
	for(int i=0;i<rl_point;i++) {
		if(copied_rl_buffer[i]==0) {
			argv[argi]=&copied_rl_buffer[current_begin];
			current_begin=i+1;
			argi++;
		}
	}
	void **ret=malloc(sizeof(void *)*2);
	ret[0]=(void *)(int64_t)(argc);
	ret[1]=argv;
	return ret;
}

void free_converted_args(void **c) {
	char **argv=(char **)c[1];
	free(argv[0]);
	free(argv);
	free(c);
}

char **fb_readline_completion(const char *text, int start, int end) {
	if(start==0) {
		tmpFunctionList=GetFunctionList();
		rl_attempted_completion_over=1;
		return rl_completion_matches(text,fb_command_generator);
	}
	if(text[0]=='-') {
		return rl_completion_matches(text,fb_args_generator);
	}
	/*unsigned int last_arg_end=0;
	unsigned int last_arg_begin=0;
	for(int point=rl_point-1;point>=0;point--) {
		if(!last_arg_end&&rl_line_buffer[point]==' ') {
			last_arg_end=point-1;
			continue;
		}
		if(last_arg_end&&!last_arg_begin&&rl_line_buffer[point]==' ') {
			last_arg_begin=point+1;
			break;
		}
	}
	char *lastarg=malloc(last_arg_end-last_arg_begin+1);
	for(int point=last_arg_begin;point<=last_arg_end;point++) {
		lastarg[point-last_arg_begin]=rl_line_buffer[point];
	}
	lastarg[last_arg_end-last_arg_begin+1]=0;
	unsigned int first_blank=0;
	for(int p=0;p<rl_point;p++) {
		if(rl_line_buffer[p]==' ')) {
			first_blank=p;
			break;
		}
	}*/
	rl_attempted_completion_over=1;
	void **args=readline_to_args();
	int argc=(int)(int64_t)args[0];
	char **argv=(char **)args[1];
	char *command_pr=argv[0];
	if(strcmp("exit",command_pr)==0||strcmp("fbexit",command_pr)==0||strcmp("logout",command_pr)==0||strcmp("lang",command_pr)==0||strcmp("ingameping",command_pr)==0||strcmp("set",command_pr)==0||strcmp("setend",command_pr)==0||strcmp("logout",command_pr)==0||strcmp("say",command_pr)==0) {
		free_converted_args(args);
		return NULL;
	}else if(strcmp("delay",command_pr)==0) {
		if(argc==1) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_delay_first_arg_generator);
		}else if(argc==2) {
			if(strcmp("mode",argv[1])==0) {
				free_converted_args(args);
				return rl_completion_matches(text,fb_set_or_get_generator);
			}
			free_converted_args(args);
			return NULL;
		}else if(argc==3) {
			if(strcmp("mode",argv[1])==0) {
				if(strcmp("set",argv[2])==0) {
					free_converted_args(args);
					return rl_completion_matches(text,fb_delay_mode_generator);
				}
			}
		}
		free_converted_args(args);
		return NULL;
	}else if(strcmp("get",command_pr)==0) {
		if(argc==1) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_get_args_generator);
		}
		free_converted_args(args);
		return NULL;
	}else if(strcmp("task",command_pr)==0) {
		if(argc==1) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_task_root_args_generator);
		}else if(argc==3&&strcmp("setdelaymode",argv[1])==0) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_delay_mode_generator);
		}
		free_converted_args(args);
		return NULL;
	}else if(strcmp("tasktype",command_pr)==0) {
		if(argc==1) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_task_type_generator);
		}
		free_converted_args(args);
		return NULL;
	}else if(strcmp("progress",command_pr)==0) {
		if(argc==1) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_bool_value_generator);
		}
		free_converted_args(args);
		return NULL;
	}else{
		if(strcmp("--path",argv[argc-1])==0||strcmp("-p",argv[argc-1])==0||strcmp("-path",argv[argc-1])==0) {
			free_converted_args(args);
			return rl_completion_matches(text,rl_filename_completion_function);
		}else if(strcmp("--facing",argv[argc-1])==0||strcmp("-f",argv[argc-1])==0||strcmp("-facing",argv[argc-1])==0) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_facing_enum_generator);
		}else if(strcmp("--shape",argv[argc-1])==0||strcmp("-s",argv[argc-1])==0||strcmp("-shape",argv[argc-1])==0) {
			free_converted_args(args);
			return rl_completion_matches(text,fb_shape_enum_generator);
		}
	}
	free_converted_args(args);
	return NULL;
}

char *fb_command_generator(const char *text, int state) {
	static int funclist_i, len;
	if(!state) {
		funclist_i=0;
		len=strlen(text);
	}
	//printf("%s\n",tmpFunctionList[funclist_i]);
	char *name;
	while((name=tmpFunctionList[funclist_i])!=NULL) {
		funclist_i++;
		/*if(!name) {
			free(tmpFunctionList);
			tmpFunctionList=NULL;
			return NULL;
		}*/
		if(strncmp(name,text,len)==0) {
			// free() will be done by readline.
			return name;
		}else{
			free(name);
		}
	}
	free(tmpFunctionList);
	tmpFunctionList=NULL;
	return NULL;
}

#define BUILD_GENERATOR(func_name, pool) static char *func_name(const char *text, int state) { \
	static int index, len; \
	if(!state) { \
		index=0; \
		len=strlen(text); \
	} \
	const char *name; \
	while((name=pool[index])!=NULL) { \
		index++; \
		if(strncmp(name,text,len)==0) { \
			char *newbuf=malloc(strlen(name)+1); \
			memcpy(newbuf, name, strlen(name)+1); \
			return newbuf; \
		} \
	} \
	return NULL; \
}

const char *flags_pool[]={
	"--assignnbtdata", "-assignnbtdata", "-nbt"
	"--excludecommands","-excludecommands",
	"--invalidatecommands","-invalidatecommands",
	"--strict","-strict","-S",
	"--length","-l","-length",
	"--width","-width","-w",
	"--height","-height","-h",
	"--radius","-radius","-r",
	"--mapX","-mapX",
	"--mapZ","-mapZ",
	"--mapY","-mapY",
	"--facing","-facing","-f",
	"--path","-path","-p",
	"--shape","-shape","-s",
	"--block","-block","-b",
	"--data","-data","-d",
	"--resume","-resume",NULL
};

BUILD_GENERATOR(fb_args_generator, flags_pool)

const char *delay_first_args[]= {
	"set",
	"mode",
	"threshold",
	NULL
};

BUILD_GENERATOR(fb_delay_first_arg_generator, delay_first_args)

const char *set_or_get_dict[]={
	"get",
	"set",
	NULL
};

BUILD_GENERATOR(fb_set_or_get_generator,set_or_get_dict);

const char *fb_delay_modes[]={
	"continuous",
	"discrete",
	"none",
	NULL
};

BUILD_GENERATOR(fb_delay_mode_generator,fb_delay_modes)

const char *fb_get_pos_dict[]={
	"begin",
	"end",
	NULL
};

BUILD_GENERATOR(fb_get_args_generator,fb_get_pos_dict)

const char *fb_task_root_args[]= {
	"list",
	"pause",
	"resume",
	"break",
	"setdelay",
	"setdelaymode",
	"setdelaythreshold",
	NULL
};

BUILD_GENERATOR(fb_task_root_args_generator,fb_task_root_args)

const char *fb_task_type_dict[]={
	"sync",
	"async",
	NULL
};

BUILD_GENERATOR(fb_task_type_generator,fb_task_type_dict)

const char *bool_value_dict[]={
	"true",
	"false",
	NULL
};

BUILD_GENERATOR(fb_bool_value_generator,bool_value_dict)

const char *fb_facing_enum[]={
	"x",
	"y",
	"z",
	NULL
};

BUILD_GENERATOR(fb_facing_enum_generator,fb_facing_enum)

const char *fb_shape_enum[]={
	"hollow",
	"solid",
	NULL
};

BUILD_GENERATOR(fb_shape_enum_generator,fb_shape_enum)
