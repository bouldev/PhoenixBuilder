#pragma once
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <memory.h>
#include <windows.h>
#include <wchar.h>
#include <unistd.h>
#include <commctrl.h>

#include "phoenixbuilder-windows-shared.h"

#define TextOutS(a,b,c,d) TextOutW(a,b,c,d,wcslen(d))

#define VERSION "0.0.1"
// FastBuilder's Main Window
extern HWND window;
#define FB_WINDOWHEIGHT 600
#define FB_WINDOWWIDTH  600
char *resolveGoStringLT(GoString str);
char *resolveGoString(GoString str);
GoString buildGoString(char *str);
void freeGoString(GoString str);

void showPanicMessage(char *msg);
void clearBoard(void);
void redrawBoard(void);

extern void (*currentPainter)(void);
extern void (*currentHandler)(WPARAM,LPARAM);
extern PAINTSTRUCT sharedPaintStruct;
extern HMODULE fastbuilder;
extern HINSTANCE globalHInstance;

// ui/welcome.c
void showWelcomeScene(void);
void showLoginFailed(char *);

// ui/main.c
void showMainScene(void);
extern void (*executeFastBuilderCommand)(const char *command);
extern void (*executeMinecraftCommand)(const char *command);
extern void (*sendChat)(const char *content);
extern char *(*getBuildPos)(void);
extern char *(*getEndPos)(void);
