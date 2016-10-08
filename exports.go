// Copyright (c) 2012-2016 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// Package gortlsdr wraps librtlsdr, which turns your Realtek RTL2832 based
// DVB dongle into a SDR receiver.
//

package rtlsdr

import (
	"unsafe"
)

/*
#include <rtl-sdr.h>
*/
import "C"

//export goRTLSDRCallback
func goRTLSDRCallback(p1 *C.uchar, p2 C.uint32_t, u unsafe.Pointer) {
	ctx := contexts.get(uint32(uintptr(u)))
	if ctx == nil {
		return
	}

	n := int(p2)
	buf := (*[1 << 24]byte)(unsafe.Pointer(p1))[:n:n]
	if ctx.clientCb2 != nil {
		ctx.clientCb2(ctx, buf, ctx.userCtx)
	} else {
		ctx.clientCb(buf)
	}
}
