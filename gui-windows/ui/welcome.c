#include "../main.h"

static BOOL _hasToken;
static HFONT bigFBFont;
static int beginXPos, bigTextWidth;
static HWND usernameEdit, passwordEdit, rentalServerCodeEdit, rentalServerPasswordEdit;
static HWND connectBtn;
static HWND removeTokenBtn;

static void paintWelcomeScene() {
	HDC hdc=BeginPaint(window, &sharedPaintStruct);
	SetBkMode(hdc, TRANSPARENT);
	int currentY=160;
	if(!_hasToken) {
		TextOutS(hdc,beginXPos,currentY,L"FastBuilder User Center's username");
		currentY+=50;
		TextOutS(hdc,beginXPos,currentY,L"FastBuilder User Center's password");
		currentY+=30;
	}else{
		currentY-=50;
	}
	currentY+=50;
	TextOutS(hdc,beginXPos,currentY,L"Rental server code");
	currentY+=50;
	TextOutS(hdc,beginXPos,currentY,L"Rental server password (leave this field blank for no password)");
	if(_hasToken) {
		currentY+=50;
		TextOutS(hdc,beginXPos,currentY,L"You don't have to enter the credential for");
		TextOutS(hdc,beginXPos,currentY+15,L"fastbuilder user center for we've");
		TextOutS(hdc,beginXPos,currentY+30,L"saved it for you, for removing it,");
		TextOutS(hdc,beginXPos,currentY+45,L"please press the button below.");
	}
	if(((char(*)())GetProcAddress(fastbuilder,"isDebugMode"))()) {
		TextOutS(hdc,-6,140,L"DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE DEBUGMODE");
	}
	SelectObject(hdc, bigFBFont);
	TextOutS(hdc,beginXPos,50,L"FastBuilder");
	EndPaint(window, &sharedPaintStruct);
}

static void disableAllControls() {
	EnableWindow(usernameEdit,0);
	EnableWindow(passwordEdit,0);
	EnableWindow(rentalServerCodeEdit,0);
	EnableWindow(rentalServerPasswordEdit,0);
	EnableWindow(connectBtn,0);
}

static void welcomeSceneInteractionHandler(WPARAM wParam, LPARAM lParam) {
	if(wParam==BN_CLICKED) {
		if((HWND)lParam==connectBtn) {
			if(((char(*)())GetProcAddress(fastbuilder,"isDebugMode"))()) {
				disableAllControls();
				void (*runLibClient)(char*,char*,char*,char*)=(void*)GetProcAddress(fastbuilder,"runLibClient");
				runLibClient(VERSION, "", "", "");
				return;
			}
			char serverCode[16]={0};
			serverCode[0]=16;
			// They SHOULD be ASCII chars so let's use "A" methods only
			int sc=SendMessageA(rentalServerCodeEdit,EM_GETLINE,0,(LPARAM)serverCode);
			char serverPassword[8]={0};
			serverPassword[0]=8;
			int sp=SendMessageA(rentalServerPasswordEdit,EM_GETLINE,0,(LPARAM)serverPassword);
			if(!sc) {
				MessageBoxW(NULL,L"You can't leave server code field empty.",L"Warning",MB_ICONERROR);
				return;
			}
			char *fbtoken;
			if(usernameEdit) {
				char username[32]={0};
				username[0]=32;
				char password[256]={0};
				password[0]=255;
				int unl=SendMessageA(usernameEdit,EM_GETLINE,0,(LPARAM)username);
				int pwl=SendMessageA(passwordEdit,EM_GETLINE,0,(LPARAM)password);
				if(!unl) {
					MessageBoxW(NULL,L"You can't leave username field empty.",L"Warning",MB_ICONERROR);
					return;
				}
				if(!pwl) {
					MessageBoxW(NULL,L"You can't leave password field empty.",L"Warning",MB_ICONERROR);
					return;
				}
				GoString Gun=buildGoString(username);
				GoString Gpw=buildGoString(password);
				char *tempToken=((char *(*)(GoString,GoString))GetProcAddress(fastbuilder,"generateTempToken"))(Gun,Gpw);
				freeGoString(Gun);
				freeGoString(Gpw);
				fbtoken=tempToken;
			}else{
				fbtoken=((char *(*)())GetProcAddress(fastbuilder,"loadToken"))();
				// Golang part will panic automatically so here is no exception checks.
			}
			disableAllControls();
			void (*runLibClient)(char*,char*,char*,char*)=(void*)GetProcAddress(fastbuilder,"runLibClient");
			runLibClient(VERSION, fbtoken, serverCode, serverPassword);
			free(fbtoken);
		}else if((HWND)lParam==removeTokenBtn){
			((void(*)())GetProcAddress(fastbuilder,"removeToken"))();
			clearBoard();
			showWelcomeScene();
			redrawBoard();
		}
	}
}

void showLoginFailed(char *msg) {
	MessageBoxA(window,msg,"Failed to login",MB_ICONERROR);
	EnableWindow(usernameEdit,TRUE);
	EnableWindow(passwordEdit,TRUE);
	EnableWindow(rentalServerCodeEdit,TRUE);
	EnableWindow(rentalServerPasswordEdit,TRUE);
	EnableWindow(connectBtn,TRUE);
	free(msg);
}

void showWelcomeScene() {
	usernameEdit=passwordEdit=rentalServerCodeEdit=rentalServerPasswordEdit=NULL;
	connectBtn=NULL;
	removeTokenBtn=NULL;
	_hasToken=(BOOL)(unsigned char(*)())GetProcAddress(fastbuilder,"hasToken")();
	printf("hasToken=%d\n",_hasToken);
	LOGFONT df;
	GetObject(GetStockObject(DEFAULT_GUI_FONT),sizeof(LOGFONT),&df);
	bigFBFont=CreateFont(df.lfHeight*6,df.lfWidth*6,df.lfEscapement,df.lfOrientation,df.lfWeight,df.lfItalic,df.lfUnderline,df.lfStrikeOut,df.lfCharSet,df.lfOutPrecision,df.lfClipPrecision,df.lfQuality,df.lfPitchAndFamily,df.lfFaceName);
	beginXPos=(FB_WINDOWWIDTH-df.lfWidth*6)/4;
	bigTextWidth=beginXPos*2;
	int currentY=180;
	if(!_hasToken) {
		usernameEdit=CreateWindowW(L"EDIT",NULL,WS_BORDER|WS_CHILD|WS_VISIBLE|ES_LEFT,beginXPos,currentY,bigTextWidth,25,window,NULL,globalHInstance,NULL);
		currentY+=50; // 25+15
		passwordEdit=CreateWindowW(L"EDIT",NULL,WS_BORDER|WS_CHILD|WS_VISIBLE|ES_LEFT|ES_PASSWORD,beginXPos,currentY,bigTextWidth,25,window,NULL,globalHInstance,NULL);
		currentY+=50;
		SendMessage(usernameEdit,EM_LIMITTEXT,31,0);
		SendMessage(passwordEdit,EM_LIMITTEXT,254,0);
		currentY+=30;
	}
	rentalServerCodeEdit=CreateWindowW(L"EDIT",NULL,WS_BORDER|WS_CHILD|WS_VISIBLE|ES_LEFT,beginXPos,currentY,bigTextWidth,25,window,NULL,globalHInstance,NULL);
	currentY+=50;
	rentalServerPasswordEdit=CreateWindowW(L"EDIT",NULL,WS_BORDER|WS_CHILD|WS_VISIBLE|ES_LEFT|ES_PASSWORD,beginXPos,currentY,bigTextWidth,25,window,NULL,globalHInstance,NULL);
	if(_hasToken) {
		currentY+=100;
		removeTokenBtn=CreateWindowW(L"BUTTON",L"Remove Token",WS_BORDER|WS_CHILD|WS_VISIBLE|ES_LEFT|ES_PASSWORD,beginXPos,currentY,120,40,window,NULL,globalHInstance,NULL);
	}
	SendMessage(rentalServerCodeEdit,EM_LIMITTEXT,15,0);
	SendMessage(rentalServerPasswordEdit,EM_LIMITTEXT,7,0);
	HFONT bigConnectFont=CreateFont(df.lfHeight*3,df.lfWidth*3,df.lfEscapement,df.lfOrientation,df.lfWeight,df.lfItalic,df.lfUnderline,df.lfStrikeOut,df.lfCharSet,df.lfOutPrecision,df.lfClipPrecision,df.lfQuality,df.lfPitchAndFamily,df.lfFaceName);
	SendMessage(connectBtn=CreateWindowW(L"BUTTON",L"Connect",WS_VISIBLE|WS_CHILD|WS_TABSTOP|BS_NOTIFY,FB_WINDOWWIDTH/4+40,FB_WINDOWHEIGHT-160,FB_WINDOWWIDTH/2-80,80,window,NULL,globalHInstance,NULL),WM_SETFONT,(WPARAM)bigConnectFont,TRUE);
	currentPainter=&paintWelcomeScene;
	currentHandler=&welcomeSceneInteractionHandler;
}