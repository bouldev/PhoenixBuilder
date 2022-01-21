FLAGS = -Wl,--enable-stdcall-fixup -luser32 -lcomdlg32 -Wno-pointer-to-int-cast -mwindows -m64 -march=x86-64 -luser32 -lkernel32 -lgdi32 -lwinmm -lcomctl32 -ladvapi32 -lshell32 -lpsapi -nodefaultlibs -nostdlib -lmsvcrt -lucrt -D_UCRT=1 -D_NO_CRT_STDIO_INLINE

all: fb.exe a

clean:
	rm fb.exe

fb.exe: main.c a
	x86_64-w64-mingw32-windres res/resources.rc -O coff -o resources.o
	x86_64-w64-mingw32-gcc resources.o main.c ui/welcome.c ui/main.c -I../build --static -Wl,-estart -o fb.exe $(FLAGS)
	rm resources.o

a:
	echo a