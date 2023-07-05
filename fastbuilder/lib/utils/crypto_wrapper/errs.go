package crypto_wrapper

import "errors"

var ErrZeroContentLength = errors.New("zero content length")
var ErrDataSizeAndKeySizeMismatch = errors.New("data size and key size mismatch, check padding")
