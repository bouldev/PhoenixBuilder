#include "../main.h"

void (*executeFastBuilderCommand)(const char *command);
void (*executeMinecraftCommand)(const char *command);
void (*sendChat)(const char *content);
char *(*getBuildPos)(void);
char *(*getEndPos)(void);

static void mainScenePainter() {
	HDC hdc=BeginPaint(window, &sharedPaintStruct);
	SetBkMode(hdc, TRANSPARENT);
	wchar_t buf[64]={0};
	swprintf(buf,64,L"Position: %S",getBuildPos());
	TextOutS(hdc,20,20,buf);
	swprintf(buf,64,L"End Position: %S",getEndPos());
	TextOutS(hdc,20,35,buf);
	EndPaint(window, &sharedPaintStruct);
}

static void mainSceneHandler(WPARAM wParam,LPARAM lParam) {
	
}

void showMainScene() {
	clearBoard();
	executeFastBuilderCommand=(void *)GetProcAddress(fastbuilder,"_executeFastBuilderCommand");
	executeMinecraftCommand=(void *)GetProcAddress(fastbuilder,"_executeMinecraftCommand");
	sendChat=(void*)GetProcAddress(fastbuilder,"_sendChat");
	getBuildPos=(void*)GetProcAddress(fastbuilder,"_getBuildPos");
	getEndPos=(void*)GetProcAddress(fastbuilder,"_getEndPos");
	currentPainter=&mainScenePainter;
	currentHandler=&mainSceneHandler;
	redrawBoard();
}