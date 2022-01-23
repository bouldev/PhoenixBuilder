package I18n

import (
	"syscall"
	"fmt"
	"os"
)

func ProcessSystemFileError(err error) error {
	finalOne:=uint16(SysError_NoTranslation)
	if(!HasTranslationFor(SysError_HasTranslation)) {
		return fmt.Errorf("%v",err)
	}
	patherror, succ:=err.(*os.PathError)
	if(!succ) {
		// Not an os.PathError
		return fmt.Errorf("%v",err)
	}
	syserr:=patherror.Err
	// Not included all errors
	switch syserr {
	case syscall.EACCES:
		finalOne=SysError_EACCES
	case syscall.EBUSY:
		finalOne=SysError_EBUSY
	case syscall.EINVAL:
		finalOne=SysError_EINVAL
	case syscall.EISDIR:
		finalOne=SysError_EISDIR
	case syscall.ENOENT:
		finalOne=SysError_ENOENT
	case syscall.ETXTBSY:
		finalOne=SysError_ETXTBSY
	}
	if(finalOne==SysError_NoTranslation) {
		return fmt.Errorf("%v",err)
	}
	return fmt.Errorf(T(SysError_HasTranslation),patherror.Path,T(finalOne))
}