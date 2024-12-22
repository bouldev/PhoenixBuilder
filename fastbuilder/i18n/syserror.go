package I18n

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"fmt"
	"os"
	"syscall"
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