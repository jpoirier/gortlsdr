// Copyright (c) 2012 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// Package gortlsdr wraps librtlsdr, which turns your Realtek RTL2832 based
// DVB dongle into a SDR receiver.
//

package rtlsdr

import (
	"reflect"
	"unsafe"
)

/*
#include <rtl-sdr.h>
*/
import "C"

//export goCallback
func goCallback(p1 *C.uchar, p2 C.uint32_t, p3 unsafe.Pointer) {
	// c buffer to go slice without copying
	var buf []byte
	length := int(p2)
	b := (*reflect.SliceHeader)((unsafe.Pointer(&buf)))
	b.Cap = length
	b.Len = length
	b.Data = uintptr(unsafe.Pointer(p1))
	clientCb(buf, (*UserCtx)(p3))
}

//export goCallback2
func goCallback2(p1 *C.uchar, p2 C.uint32_t, p3 unsafe.Pointer) {
	// c buffer to go slice without copying
	var buf []byte
	length := int(p2)
	b := (*reflect.SliceHeader)((unsafe.Pointer(&buf)))
	b.Cap = length
	b.Len = length
	b.Data = uintptr(unsafe.Pointer(p1))

	if c, ok := (*(*UserCtx)(p3)).(*CustUserCtx); ok {
		c.ClientCb(buf, c.Userctx)
	} else {
		clientCb(buf, (*UserCtx)(p3))
	}
}
