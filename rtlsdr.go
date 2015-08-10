// Copyright (c) 2012 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// rtlsdr wraps librtlsdr, which turns your DVB dongle into a SDR receiver.
//

package rtlsdr

import (
	"errors"
	"unsafe"
)

/*
// On linux, you may need to unload the kernel DVB driver via:
//     $ sudo rmmod dvb_usb_rtl28xxu rtl2832
// If building libusb from source, to regenerate the configure file use:
//     $ autoreconf -fvi

#cgo linux LDFLAGS: -lrtlsdr
#cgo darwin LDFLAGS: -lrtlsdr
#cgo windows CFLAGS: -IC:/WINDOWS/system32
#cgo windows LDFLAGS: -lrtlsdr -LC:/WINDOWS/system32

#include <stdlib.h>
#include <rtl-sdr.h>

extern void goCallback(unsigned char *buf, uint32_t len, void *ctx);
rtlsdr_read_async_cb_t get_go_cb() {
	return (rtlsdr_read_async_cb_t)goCallback;
}
*/
import "C"

// Current version.
var PackageVersion = "v2.5"

// Device context.
type Context struct {
	dev *C.rtlsdr_dev_t
}

// Per user device context.
type UserCtx interface{}

// Sampling mode type.
type SamplingMode int

// These constants are used to set default parameter values.
const (
	DefaultGAIN           = "auto"
	DefaultFc             = 80e6
	DefaultRs             = 1.024e6
	DefaultReadSize       = 1024
	CrystalFreq           = 28800000
	DefaultSampleRate     = 2048000
	DefaultAsyncBufNumber = 32
	DefaultBufLength      = (16 * 16384)
	MinimalBufLength      = 512
	MaximalBufLength      = (256 * 16384)
)

const (
	libusbSuccess = iota * -1
	libusbErrorIo
	libusbErrorInvalidParam
	libusbErrorAccess
	libusbErrorNoDevice
	libusbErrorNotFound
	libusbErrorBusy
	libusbErrorTimeout
	libusbErrorOverflow
	libusbErrorPipe
	libusbErrorInterrupted
	libusbErrorNoMem
	libusbErrorNotSupported
	libusbErrorOther = -99
)

// Sampling modes.
const (
	SamplingNone SamplingMode = iota
	SamplingIADC
	SamplingQADC
	SamplingUnknown
)

var errMap = map[int]error{
	libusbSuccess:           nil,
	libusbErrorIo:           errors.New("input/output error"),
	libusbErrorInvalidParam: errors.New("invalid parameter(s)"),
	libusbErrorAccess:       errors.New("access denied (insufficient permissions)"),
	libusbErrorNoDevice:     errors.New("no such device (it may have been disconnected)"),
	libusbErrorNotFound:     errors.New("entity not found"),
	libusbErrorBusy:         errors.New("resource busy"),
	libusbErrorTimeout:      errors.New("operation timed out"),
	libusbErrorOverflow:     errors.New("overflow"),
	libusbErrorPipe:         errors.New("pipe error"),
	libusbErrorInterrupted:  errors.New("system call interrupted (perhaps due to signal)"),
	libusbErrorNoMem:        errors.New("insufficient memory"),
	libusbErrorNotSupported: errors.New("operation not supported or unimplemented on this platform"),
	libusbErrorOther:        errors.New("unknown error"),
}

// Sampling modes map.
var SamplingModes = map[SamplingMode]string{
	SamplingNone:    "Disabled",
	SamplingIADC:    "I-ADC Enabled",
	SamplingQADC:    "Q-ADC Enabled",
	SamplingUnknown: "Unknown",
}

var tunerTypes = map[uint32]string{
	C.RTLSDR_TUNER_UNKNOWN: "RTLSDR_TUNER_UNKNOWN",
	C.RTLSDR_TUNER_E4000:   "RTLSDR_TUNER_E4000",
	C.RTLSDR_TUNER_FC0012:  "RTLSDR_TUNER_FC0012",
	C.RTLSDR_TUNER_FC0013:  "RTLSDR_TUNER_FC0013",
	C.RTLSDR_TUNER_FC2580:  "RTLSDR_TUNER_FC2580",
	C.RTLSDR_TUNER_R820T:   "RTLSDR_TUNER_R820T",
	C.RTLSDR_TUNER_R828D:   "RTLSDR_TUNER_R828D",
}

type ReadAsyncCb_T func([]byte, *UserCtx)

var clientCb ReadAsyncCb_T

// libusbError returns a textual error desciption from errno.
func libusbError(errno int) error {
	if err, ok := errMap[errno]; ok {
		return err
	}
	return errors.New("unknown error")
}

// GetDeviceCount returns the number of devices detected.
func GetDeviceCount() (count int) {
	return int(C.rtlsdr_get_device_count())
}

// GetDeviceName returns the name of the device by index.
func GetDeviceName(index int) (name string) {
	return C.GoString(C.rtlsdr_get_device_name(C.uint32_t(index)))
}

// GetDeviceUsbStrings returns the information of a device by index.
func GetDeviceUsbStrings(index int) (manufact, product, serial string, err error) {
	m := make([]byte, 257) // includes space for NULL byte
	p := make([]byte, 257)
	s := make([]byte, 257)
	i := int(C.rtlsdr_get_device_usb_strings(C.uint32_t(index),
		(*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])),
		(*C.char)(unsafe.Pointer(&s[0]))))
	return string(m), string(p), string(s), libusbError(i)
}

// GetIndexBySerial returns a device index by serial id.
func GetIndexBySerial(serial string) (index int, err error) {
	cstring := C.CString(serial)
	defer C.free(unsafe.Pointer(cstring))
	index = int(C.rtlsdr_get_index_by_serial(cstring))
	switch {
	case index >= 0:
		err = nil
	case index == -1:
		err = errors.New("serial blank")
	case index == -2:
		err = errors.New("no devices were found")
	case index == -3:
		err = errors.New("no device found matching name")
	default:
		err = errors.New("unknown error")
	}
	return
}

// Open returns an opened device by index.
func Open(index int) (c *Context, err error) {
	c = &Context{}
	i := int(C.rtlsdr_open((**C.rtlsdr_dev_t)(&c.dev),
		C.uint32_t(index)))
	return c, libusbError(i)
}

// Close closes the device.
func (c *Context) Close() (err error) {
	i := int(C.rtlsdr_close((*C.rtlsdr_dev_t)(c.dev)))
	return libusbError(i)
}

// configuration functions

// SetXtalFreq sets the crystal oscillator frequencies.
//
// Typically both ICs use the same clock. Changing the clock may make sense if
// you are applying an external clock to the tuner or to compensate the
// frequency (and samplerate) error caused by the original (cheap) crystal.
//
// Note, call this function only if you fully understand the implications.
func (c *Context) SetXtalFreq(rtlFreqHz, tunerFreqHz int) (err error) {
	i := int(C.rtlsdr_set_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rtlFreqHz),
		C.uint32_t(tunerFreqHz)))
	return libusbError(i)
}

// GetXtalFreq returns the crystal oscillator frequencies.
// Typically both ICs use the same clock.
func (c *Context) GetXtalFreq() (rtlFreqHz, tunerFreqHz int, err error) {
	i := int(C.rtlsdr_get_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint32_t)(unsafe.Pointer(&rtlFreqHz)),
		(*C.uint32_t)(unsafe.Pointer(&tunerFreqHz))))
	return rtlFreqHz, tunerFreqHz, libusbError(i)
}

// GetUsbStrings returns the device information. Note, strings may be empty.
func (c *Context) GetUsbStrings() (manufact, product, serial string, err error) {
	m := make([]byte, 257) // includes space for NULL byte
	p := make([]byte, 257)
	s := make([]byte, 257)
	i := int(C.rtlsdr_get_usb_strings((*C.rtlsdr_dev_t)(c.dev),
		(*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])),
		(*C.char)(unsafe.Pointer(&s[0]))))
	return string(m), string(p), string(s), libusbError(i)
}

// WriteEeprom writes data to the EEPROM.
func (c *Context) WriteEeprom(data []uint8, offset uint8, leng uint16) (err error) {
	i := int(C.rtlsdr_write_eeprom((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.uint8_t(offset),
		C.uint16_t(leng)))
	switch i {
	case 0:
		err = nil
	case -1:
		err = errors.New("device handle is invalid")
	case -2:
		err = errors.New("EEPROM size exceeded")
	case -3:
		err = errors.New("no EEPROM was found")
	default:
		err = errors.New("unknown error")
	}
	return
}

// ReadEeprom returns data read from the EEPROM.
func (c *Context) ReadEeprom(data []uint8, offset uint8, leng uint16) (err error) {
	i := int(C.rtlsdr_read_eeprom((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.uint8_t(offset),
		C.uint16_t(leng)))
	switch i {
	case 0:
		err = nil
	case -1:
		err = errors.New("device handle is invalid")
	case -2:
		err = errors.New("EEPROM size exceeded")
	case -3:
		err = errors.New("no EEPROM was found")
	default:
		err = errors.New("unknown error")
	}
	return
}

// SetCenterFreq sets the center frequency.
func (c *Context) SetCenterFreq(freqHz int) (err error) {
	i := int(C.rtlsdr_set_center_freq((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(freqHz)))
	return libusbError(i)
}

// GetCenterFreq returns the tuned frequency.
func (c *Context) GetCenterFreq() (freqHz int) {
	return int(C.rtlsdr_get_center_freq((*C.rtlsdr_dev_t)(c.dev)))
}

// SetFreqCorrection sets the frequency correction.
func (c *Context) SetFreqCorrection(freqHz int) (err error) {
	i := int(C.rtlsdr_set_freq_correction((*C.rtlsdr_dev_t)(c.dev),
		C.int(freqHz)))
	return libusbError(i)
}

// GetFreqCorrection returns the frequency correction value.
func (c *Context) GetFreqCorrection() (ppm int) {
	return int(C.rtlsdr_get_freq_correction((*C.rtlsdr_dev_t)(c.dev)))
}

// GetTunerType returns the tuner type.
func (c *Context) GetTunerType() (tunerType string) {
	t := C.rtlsdr_get_tuner_type((*C.rtlsdr_dev_t)(c.dev))
	if tt, ok := tunerTypes[t]; ok {
		tunerType = tt
	} else {
		tunerType = "UNKNOWN"
	}
	return
}

// GetTunerGains returns a list of supported tuner gains.
// Values are in tenths of dB, e.g. 115 means 11.5 dB.
func (c *Context) GetTunerGains() (gainsTenthsDb []int, err error) {
	//	count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev), nil))
	i := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev),
		(*C.int)(unsafe.Pointer(uintptr(0)))))
	if i < 0 {
		return gainsTenthsDb, libusbError(i)
	} else if i == 0 {
		return gainsTenthsDb, nil
	}
	gainsTenthsDb = make([]int, i)
	C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev),
		(*C.int)(unsafe.Pointer(&gainsTenthsDb[0])))
	return gainsTenthsDb, nil
}

// SetTunerGain sets the tuner gain. Note, manual gain mode
// must be enabled for this to work. Valid gain values may be
// queried using GetTunerGains.
//
// Valid values (in tenths of a dB) are:
// -10, 15, 40, 65, 90, 115, 140, 165, 190, 215, 240, 290,
// 340, 420, 430, 450, 470, 490
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (c *Context) SetTunerGain(gainsTenthsDb int) (err error) {
	i := int(C.rtlsdr_set_tuner_gain((*C.rtlsdr_dev_t)(c.dev),
		C.int(gainsTenthsDb)))
	return libusbError(i)
}

// GetTunerGain returns the tuner gain.
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (c *Context) GetTunerGain() (gainsTenthsDb int) {
	return int(C.rtlsdr_get_tuner_gain((*C.rtlsdr_dev_t)(c.dev)))
}

// SetTunerIfGain sets the intermediate frequency gain.
//
// Intermediate frequency gain stage number 1 to 6.
// Gain values are in tenths of dB, e.g. -30 means -3.0 dB.
func (c *Context) SetTunerIfGain(stage, gainsTenthsDb int) (err error) {
	i := int(C.rtlsdr_set_tuner_if_gain((*C.rtlsdr_dev_t)(c.dev),
		C.int(stage),
		C.int(gainsTenthsDb)))
	return libusbError(i)
}

// SetTunerGainMode sets the gain mode (automatic/manual).
// Manual gain mode must be enabled for the gain setter function to work.
func (c *Context) SetTunerGainMode(manualMode bool) (err error) {
	mode := 0
	if manualMode {
		mode = 1
	}
	i := int(C.rtlsdr_set_tuner_gain_mode((*C.rtlsdr_dev_t)(c.dev),
		C.int(mode)))
	return libusbError(i)
}

// SetSampleRate sets the sample rate.
//
// When applicable, the baseband filters are also selected based
// on the requested sample rate.
func (c *Context) SetSampleRate(rateHz int) (err error) {
	i := int(C.rtlsdr_set_sample_rate((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rateHz)))
	return libusbError(i)

}

// GetSampleRate returns the sample rate.
func (c *Context) GetSampleRate() (rateHz int) {
	return int(C.rtlsdr_get_sample_rate((*C.rtlsdr_dev_t)(c.dev)))
}

// SetTestMode sets to test mode.
//
// Test mode returns 8 bit counters instead of samples. Note,
// the counter is generated inside the device.
func (c *Context) SetTestMode(testMode bool) (err error) {
	mode := 0
	if testMode {
		mode = 1
	}
	i := int(C.rtlsdr_set_testmode((*C.rtlsdr_dev_t)(c.dev),
		C.int(mode)))
	return libusbError(i)
}

// SetAgcMode sets the AGC mode.
func (c *Context) SetAgcMode(AGCMode bool) (err error) {
	mode := 0
	if AGCMode {
		mode = 1
	}
	i := int(C.rtlsdr_set_agc_mode((*C.rtlsdr_dev_t)(c.dev),
		C.int(mode)))
	return libusbError(i)
}

// SetDirectSampling sets the direct sampling mode.
//
// When enabled, the IF mode of the device is activated, and
// SetCenterFreq() will control the IF-frequency of the DDC, which
// can be used to tune from 0 to 28.8 MHz (xtal frequency of the device).
func (c *Context) SetDirectSampling(mode SamplingMode) (err error) {
	i := int(C.rtlsdr_set_direct_sampling((*C.rtlsdr_dev_t)(c.dev),
		C.int(mode)))
	return libusbError(i)
}

// GetDirectSampling returns the state of direct sampling mode.
func (c *Context) GetDirectSampling() (mode SamplingMode, err error) {
	i := int(C.rtlsdr_get_direct_sampling((*C.rtlsdr_dev_t)(c.dev)))
	switch i {
	case -1:
		err = errors.New("error getting sampling mode")
	case 0:
		mode = SamplingNone
		err = nil
	case 1:
		mode = SamplingIADC
		err = nil
	case 2:
		mode = SamplingQADC
		err = nil
	default:
		mode = SamplingUnknown
		err = errors.New("unknown sampling mode state")
	}
	return
}

// SetOffsetTuning sets the offset tuning mode for zero-IF tuners, which
// avoids problems caused by the DC offset of the ADCs and 1/f noise.
func (c *Context) SetOffsetTuning(enable bool) (err error) {
	mode := 1
	if !enable {
		mode = 0
	}
	i := int(C.rtlsdr_set_offset_tuning((*C.rtlsdr_dev_t)(c.dev), C.int(mode)))
	return libusbError(i)
}

// GetOffsetTuning returns the offset tuning mode.
func (c *Context) GetOffsetTuning() (enabled bool, err error) {
	i := int(C.rtlsdr_get_offset_tuning((*C.rtlsdr_dev_t)(c.dev)))
	switch i {
	case -1:
		err = errors.New("error getting offset tuning mode")
	case 0:
		enabled = false
		err = nil
	case 1:
		enabled = true
		err = nil
	default:
		err = errors.New("unknown offset tuning mode state")
	}
	return
}

// streaming functions

// ResetBuffer resets the streaming buffer.
func (c *Context) ResetBuffer() (err error) {
	i := int(C.rtlsdr_reset_buffer((*C.rtlsdr_dev_t)(c.dev)))
	return libusbError(i)
}

// ReadSync performs a synchronous read of samples and returns
// the number of samples read.
func (c *Context) ReadSync(buf []uint8, leng int) (nRead int, err error) {
	i := int(C.rtlsdr_read_sync((*C.rtlsdr_dev_t)(c.dev),
		unsafe.Pointer(&buf[0]),
		C.int(leng),
		(*C.int)(unsafe.Pointer(&nRead))))
	return nRead, libusbError(i)
}

// ReadAsync reads samples asynchronously. Note, this function
// will block until canceled using CancelAsync
//
// Optional bufNum buffer count, bufNum * bufLen = overall buffer size,
// set to 0 for default buffer count (32).
// Optional bufLen buffer length, must be multiple of 512, set to 0 for
// default buffer length (16 * 32 * 512).
func (c *Context) ReadAsync(f ReadAsyncCb_T, userctx *UserCtx, bufNum,
	bufLen int) (err error) {
	clientCb = f
	i := int(C.rtlsdr_read_async((*C.rtlsdr_dev_t)(c.dev),
		(C.rtlsdr_read_async_cb_t)(C.get_go_cb()),
		unsafe.Pointer(userctx),
		C.uint32_t(bufNum),
		C.uint32_t(bufLen)))
	return libusbError(i)
}

// CancelAsync cancels all pending asynchronous operations.
func (c *Context) CancelAsync() (err error) {
	i := int(C.rtlsdr_cancel_async((*C.rtlsdr_dev_t)(c.dev)))
	return libusbError(i)
}
