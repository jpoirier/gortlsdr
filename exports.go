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

//export go_callback
func go_callback(p1 *C.char, p2 C.uint32_t, p3 unsafe.Pointer) {
	// c buffer to go slice without copying
	var buf []int8
	length := int(p2)
	b := (*reflect.SliceHeader)((unsafe.Pointer(&buf)))
	b.Cap = length
	b.Len = length
	b.Data = uintptr(unsafe.Pointer((*int8)(p1)))
	clientCb(buf)
}

