#include "main.h"
#include <stdint.h>

GoString FBVersion;
HMODULE fastbuilder;

void (*currentPainter)(void)=NULL;
void (*currentHandler)(WPARAM,LPARAM)=NULL;
PAINTSTRUCT sharedPaintStruct;
HINSTANCE globalHInstance;
HWND window;

char *resolveGoStringLT(GoString str) {
	char *ret=malloc(str.n+1);
	memcpy(ret,str.p,str.n);
	ret[str.n]=0;
	return ret;
}

char *resolveGoString(GoString str) {
	static char ret[1024*5]={0};
	// use static so it'll be a pre-allocated space and the content will be
	// replaced every time calling this function.
	memcpy(ret,str.p,str.n);
	ret[str.n]=0;
	return ret;
}

GoString buildGoString(char *str) {
	int len=strlen(str);
	GoString out;
	char *dyn=malloc(len);
	// For GoString, len+1 isn't a requirement.
	memcpy(dyn,str,len);
	out.p=dyn;
	out.n=len;
	return out;
}

void freeGoString(GoString str) {
	free((void*)str.p);
}

void showPanicMessage(char *msg) {
	MessageBoxA(window,msg,"Panic",MB_ICONERROR);
	free(msg);
	exit(1);
}

void redrawBoard() {
	RedrawWindow(window,NULL,NULL,RDW_INTERNALPAINT|RDW_UPDATENOW|RDW_INVALIDATE|RDW_ERASENOW|RDW_ERASE);
}

BOOL CALLBACK boardCleaner(HWND w,LPARAM lp) {
	SendMessageA(w,WM_NCDESTROY,0,0);
	return TRUE;
}

void clearBoard() {
	EnumChildWindows(window,boardCleaner,0);
	currentPainter=NULL;
	currentHandler=NULL;
	redrawBoard();
}

void initFinished() {
	
}

void *fbcallbacks[] = {
	showPanicMessage,
	showLoginFailed,
	showMainScene
};

static LRESULT CALLBACK mainWindowProc(HWND hwnd,UINT msg,WPARAM wParam,LPARAM lParam) {
	if(msg==WM_PAINT) {
		if(currentPainter)currentPainter(/*hwnd=window so no need to send it*/);
	}else if(msg==WM_CLOSE) {
		exit(0);
	}else if(msg==WM_DESTROY) {
		exit(0);
	}else if(msg==WM_COMMAND) {
		if(currentHandler)currentHandler(wParam,lParam);
	}
	return DefWindowProc(hwnd,msg,wParam,lParam);
}

void WINAPI start() {
	printf("FastBuilder is loading, please wait patiently...\n");
	fastbuilder=LoadLibraryA("phoenixbuilder-windows-shared.dll");
	if(!fastbuilder) {
		MessageBoxA(NULL,"Cannot load fastbuilder library.","Fatal error",MB_ICONERROR);
		exit(12);
		return;
	}
	//printf("start();\n");
	GoString (*getFBVersion)(void)=(void *)GetProcAddress(fastbuilder,"GetFBVersion");
	FBVersion=getFBVersion();
	printf("FastBuilder GUI (under Windows)\nCore Version: %s\nGUI Version: " VERSION "\nAuthor: Ruphane\n\n",resolveGoString(FBVersion));
	InitCommonControls();
	WNDCLASSEXW windowclass={0};
	windowclass.cbSize=sizeof(WNDCLASSEXW);
	windowclass.lpfnWndProc=mainWindowProc;
	windowclass.lpszClassName=L"FBGUIMain";
	windowclass.style=CS_HREDRAW|CS_VREDRAW;
	windowclass.hbrBackground=(HBRUSH)(COLOR_3DFACE+1);
	windowclass.hInstance=GetModuleHandle(NULL);
	windowclass.hCursor=LoadCursor(NULL,IDC_ARROW);
	RegisterClassExW(&windowclass);
	RECT mwrct={0};
	mwrct.right=FB_WINDOWWIDTH;
	mwrct.bottom=FB_WINDOWHEIGHT;
	AdjustWindowRect(&mwrct, WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX, FALSE);
	window=CreateWindowExW(0,L"FBGUIMain",L"FastBuilder",WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX,50,50,mwrct.right-mwrct.left,mwrct.bottom-mwrct.top,NULL,NULL,GetModuleHandle(NULL),NULL);
	globalHInstance=(HINSTANCE)GetWindowLongPtrW(window,GWLP_HINSTANCE);
	((void(*)(void*))GetProcAddress(fastbuilder,"setCallbacks"))(fbcallbacks);
	showWelcomeScene();
	UpdateWindow(window);
	ShowWindow(window,SW_SHOW);
	MSG msg;
	while(GetMessage(&msg, NULL, 0, 0)) {
		TranslateMessage(&msg);
		DispatchMessage(&msg);
	}
}
