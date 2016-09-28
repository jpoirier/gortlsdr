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
	n := int(p2)
	buf := (*[1 << 24]byte)(unsafe.Pointer(p1))[:n:n]
	dev := contexts[int(uintptr(u))]
	if dev == nil {
		return
	}
	// FIXME add user context
	// TODO add device ? many advaced rtl-sdr software use device inside callback
	dev.clientCb(buf)
}
