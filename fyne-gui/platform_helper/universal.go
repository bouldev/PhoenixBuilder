package platform_helper

// void initStdoutRedirector(char *rootPath);
import "C"

func InitStdoutRedirector(rootPath string) {
	C.initStdoutRedirector(C.CString(rootPath))
}