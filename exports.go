// Copyright (c) 2012-2016 Joseph D Poirier
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

//export goRTLSDRCallback
func goRTLSDRCallback(p1 *C.uchar, p2 C.uint32_t, _ unsafe.Pointer) {
	// c buffer to go slice without copying
	var buf []byte
	length := int(p2)
	b := (*reflect.SliceHeader)((unsafe.Pointer(&buf)))
	b.Cap = length
	b.Len = length
	b.Data = uintptr(unsafe.Pointer(p1))
	clientCb(buf)
}
