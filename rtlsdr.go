// Copyright (c) 2012 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// Package gortlsdr wraps rtl-sdr, which turns your Realtek RTL2832 based
// DVB dongle into a SDR receiver
//
// The package is low level and, for the most part, is one-to-one with the
// exported C functions it wraps. Clients would typically build instrument
// drivers around the package but it can also be used directly.
//
// Lots of miscellaneous NI-488.2 information:
//     http://sdr.osmocom.org/trac/wiki/rtl-sdr
//
//
// Direct download: http://github.com/jpoirier/gortlsdr
//
package rtlsdr

// #cgo LDFLAGS: -lrtlsdr
// #include <rtl-sdr.h>
/*
extern void go_callback(char* p1, uint32_t p2, void* p3);
*/
import "C"

import (
	"log"
	// "reflect"
	"unsafe"
)

var PackageVersion string = "v0.1"

/*
TODO
----
The rtlsdr library returns inconsistent error codes:
	o
	o rtlsdr_read_sync - returns libusb error codes but ut steps on one of them (-1)
	o rtlsdr_set_xtal_freq - returns libusb error codes but steps on some of them (-1, -2, -3)
	o rtlsdr_get_center_freq - returns 0 on error, should be non-zero, e.g. -1?
	o rtlsdr_get_tuner_gain - returns 0 on error, should be non-zero, e.g. -1?
	o rtlsdr_get_sample_rate - returns 0 on error, should be non-zero, e.g. -1?
	o rtlsdr_set_testmode - returns libusb error codes but ut steps on one of them (-1)
	o rtlsdr_get_freq_correction - returns 0 on error, should be non-zero, e.g. -1?
	o
	o functions that return libusb_control_transfer's return value should state
	  error values as less than zero
	o rtlsdr_set_center_freq - rtlsdr_set_if_freq returns negative values on error
	  or zero or more bytes written (it calls libusb_control_transfer) on success
	  and dev->tuner->set_freq returns a zero on success, but rtlsdr_set_center_freq
	  performs a 'not true' check prior to setting dev->freq, which incorrectly
	  causes the freq to not get set when direct_sampling is being performed.

	GetCenterFreq
	GetFreqCorrection
	GetTunerGain
	GetSampleRate
*/

type Context struct {
	dev *C.rtlsdr_dev_t
}

type UserCtx interface{}

const (
	/* tuner types */
	TunerUnknown = C.RTLSDR_TUNER_UNKNOWN
	TunerE4000   = C.RTLSDR_TUNER_E4000
	TunerFC0012  = C.RTLSDR_TUNER_FC0012
	TunerFC0013  = C.RTLSDR_TUNER_FC0013
	TunerFC2580  = C.RTLSDR_TUNER_FC2580
	TunerR820T   = C.RTLSDR_TUNER_R820T

	/* default parameter settings */
	Default_GAIN    = "auto"
	DefaultFc       = 80e6
	DefaultRs       = 1.024e6
	DefaultReadSize = 1024
	CrystalFreq     = 28800000

	DefaultSampleRate     = 2048000
	DefaultAsyncBufNumber = 32
	DefaultBufLength      = (16 * 16384)
	MinimalBufLength      = 512
	MaximalBufLength      = (256 * 16384)
	Success               = 0
	Error                 = -1

	LibusbSuccess           = 0
	LibusbErrorIo           = -1
	LibusbErrorInvalidParam = -2
	LibusbErrorAccess       = -3
	LibusbErrorNoDevice     = -4
	LibusbErrorNotFound     = -5
	LibusbErrorBusy         = -6
	LibusbErrorTimeout      = -7
	LibusbErrorOverflow     = -8
	LibusbErrorPipe         = -9
	LibusbErrorInterrupted  = -10
	LibusbErrorNoMem        = -11
	LibusbErrorNotSupported = -12
	LibusbErrorOther        = -99
)

var TypeMap = map[int]string{
	TunerUnknown: "RTLSDR_TUNER_UNKNOWN",
	TunerE4000:   "RTLSDR_TUNER_E4000",
	TunerFC0012:  "RTLSDR_TUNER_FC0012",
	TunerFC0013:  "RTLSDR_TUNER_FC0013",
	TunerFC2580:  "RTLSDR_TUNER_FC2580",
	TunerR820T:   "RTLSDR_TUNER_R820T",
}

//type ReadAsyncCb_T func([]int8, UserCtx)
type ReadAsyncCb_T func(*int8, uint32, *UserCtx)

var clientCb ReadAsyncCb_T
var clientCtx UserCtx

//export go_callback
func go_callback(p1 *C.char, p2 C.uint32_t, p3 unsafe.Pointer) {
	// c buffer to go slice without copying
	// var buf []int8
	// length := int(p2)
	// b := (*reflect.SliceHeader)((unsafe.Pointer(&buf)))
	// b.Cap = length
	// b.Len = length
	// b.Data = uintptr(unsafe.Pointer((*int8)(p1)))
	// clientCb(buf, clientCtx)
	//clientCb(buf, clientCtx)
	clientCb((*int8)(p1), uint32(p2), (*UserCtx)(p3))
	log.Printf("called client cb...\n")
}

var GoCallback = go_callback

// GetDeviceCount gets the number of valid USB dongles detected.
//
// uint32_t rtlsdr_get_device_count(void);
func GetDeviceCount() (count int) {
	return int(C.rtlsdr_get_device_count())

}

// GetDeviceName gets the name of the USB dongle device via index,
// e.g. from an index returned from calling GainValues.
//
// const char* rtlsdr_get_device_name(uint32_t index);
func GetDeviceName(index int) (name string) {
	return C.GoString(C.rtlsdr_get_device_name(C.uint32_t(index)))
}

// Get USB device strings.
//
// int rtlsdr_get_device_usb_strings(uint32_t index, char *manufact, char *product, char *serial);
// rtlsdr_get_device_usb_strings returns 0 on success
func GetDeviceUsbStrings(index int) (manufact, product, serial string, err int) {
	m := make([]byte, 257) // includes space for NULL byte
	p := make([]byte, 257)
	s := make([]byte, 257)
	err = int(C.rtlsdr_get_device_usb_strings(C.uint32_t(index),
		(*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])),
		(*C.char)(unsafe.Pointer(&s[0]))))
	return string(m), string(p), string(s), err
}

// Open returns a valid device's context.
//
// int rtlsdr_open(rtlsdr_dev_t **dev, uint32_t index);
// rtlsdr_open returns 0 on success
func Open(index int) (c *Context, err int) {
	c = &Context{}
	err = int(C.rtlsdr_open((**C.rtlsdr_dev_t)(&c.dev),
		C.uint32_t(index)))
	return
}

// Close closes a previously opened device context.
//
// int rtlsdr_close(rtlsdr_dev_t *dev);
// rtlsdr_close returns 0 on success
func (c *Context) Close() (err int) {
	return int(C.rtlsdr_close((*C.rtlsdr_dev_t)(c.dev)))
}

// configuration functions

// Set crystal oscillator frequencies used for the RTL2832 and the tuner IC.
//
// Usually both ICs use the same clock. Changing the clock may make sense if
// you are applying an external clock to the tuner or to compensate the
// frequency (and samplerate) error caused by the original (cheap) crystal.
//
// NOTE: Call this function only if you fully understand the implications.
// Values are in Hz.
//
// int rtlsdr_set_xtal_freq(rtlsdr_dev_t *dev, uint32_t rtl_freq, uint32_t tuner_freq);
// rtlsdr_set_xtal_freq returns 0 on success
func (c *Context) SetXtalFreq(rtl_freq, tuner_freq int) (err int) {
	return int(C.rtlsdr_set_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rtl_freq),
		C.uint32_t(tuner_freq)))
}

// Get crystal oscillator frequencies used for the RTL2832 and the tuner IC.
//
// Usually both ICs use the same clock.
// Frequency values are in Hz.
//
// int rtlsdr_get_xtal_freq(rtlsdr_dev_t *dev, uint32_t *rtl_freq, uint32_t *tuner_freq);
// rtlsdr_get_xtal_freq returns 0 on success
func (c *Context) GetXtalFreq() (rtl_freq, tuner_freq, err int) {
	err = int(C.rtlsdr_get_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint32_t)(unsafe.Pointer(&rtl_freq)),
		(*C.uint32_t)(unsafe.Pointer(&tuner_freq))))
	return
}

// Get USB device strings.
//
// int rtlsdr_get_usb_strings(rtlsdr_dev_t *dev, char *manufact, char *product, char *serial);
// rtlsdr_get_usb_strings returns 0 on success
func (c *Context) GetUsbStrings() (manufact, product, serial string, err int) {
	m := make([]byte, 257) // includes space for NULL byte
	p := make([]byte, 257)
	s := make([]byte, 257)
	err = int(C.rtlsdr_get_usb_strings((*C.rtlsdr_dev_t)(c.dev),
		(*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])),
		(*C.char)(unsafe.Pointer(&s[0]))))
	return string(m), string(p), string(s), err
}

// Set the device center frequency.
//
// int rtlsdr_set_center_freq(rtlsdr_dev_t *dev, uint32_t freq);
// rtlsdr_set_center_freq returns a negative value on error
func (c *Context) SetCenterFreq(freq int) (err int) {
	return int(C.rtlsdr_set_center_freq((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(freq)))
}

// Get actual frequency the device is tuned to.
//
// Frequency values are in Hz.
//
// uint32_t rtlsdr_get_center_freq(rtlsdr_dev_t *dev);
// rtlsdr_get_center_freq returns frequency in Hz or 0 on error
func (c *Context) GetCenterFreq() (freq, err int) {
	freq = int(C.rtlsdr_get_center_freq((*C.rtlsdr_dev_t)(c.dev)))
	if freq == 0 {
		err = -1
	}
	return
}

// Set the frequency correction value for the device.
//
// Frequency values are in Hz.
//
// int rtlsdr_set_freq_correction(rtlsdr_dev_t *dev, int ppm);
// rtlsdr_set_freq_correction returns 0 on success
func (c *Context) SetFreqCorrection(ppm int) (err int) {
	return int(C.rtlsdr_set_freq_correction((*C.rtlsdr_dev_t)(c.dev),
		C.int(ppm)))
}

// Get actual frequency correction value of the device.
//
// Correction value in ppm (parts per million).
//
// int rtlsdr_get_freq_correction(rtlsdr_dev_t *dev);
// rtlsdr_get_freq_correction returns 0 on error
func (c *Context) GetFreqCorrection() (freq, err int) {
	freq = int(C.rtlsdr_get_freq_correction((*C.rtlsdr_dev_t)(c.dev)))
	if freq == 0 {
		err = -1
	}
	return
}

// Get the tuner type.
//
// enum rtlsdr_tuner rtlsdr_get_tuner_type(rtlsdr_dev_t *dev);
// rtlsdr_get_tuner_type returns enum type or unknown
func (c *Context) GetTunerType() (rtlsdr_tuner int) {
	return int(C.rtlsdr_get_tuner_type((*C.rtlsdr_dev_t)(c.dev)))
}

// Get a list of gains supported by the tuner.
//
// A NULL second parameter requests the gain count.
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
//
// int rtlsdr_get_tuner_gains(rtlsdr_dev_t *dev, int *gains);
// rtlsdr_get_tuner_gains returns <= 0 on error or the number of available gain
func (c *Context) GetTunerGains() (gains []int, err int) {
	//	count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev), nil))
	count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev),
		(*C.int)(unsafe.Pointer(uintptr(0)))))
	if count < 0 {
		err = count
		return
	} else if count == 0 {
		return
	}
	gains = make([]int, count)
	C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev),
		(*C.int)(unsafe.Pointer(&gains[0])))
	return
}

// Set the gain for the device.
// Manual gain mode must be enabled for this to work.
//
// Valid gain values (in tenths of a dB) for the E4000 tuner:
// -10, 15, 40, 65, 90, 115, 140, 165, 190, 215, 240, 290,
// 340, 420, 430, 450, 470, 490
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
//
// int rtlsdr_set_tuner_gain(rtlsdr_dev_t *dev, int gain);
// rtlsdr_set_tuner_gain returns 0 on success
func (c *Context) SetTunerGain(gain int) (err int) {
	return int(C.rtlsdr_set_tuner_gain((*C.rtlsdr_dev_t)(c.dev),
		C.int(gain)))
}

// Get actual gain the device is configured to.
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
//
// int rtlsdr_get_tuner_gain(rtlsdr_dev_t *dev);
// rtlsdr_get_tuner_gain returns 0 on error or the gain in tenths of a dB
func (c *Context) GetTunerGain() (gain, err int) {
	gain = int(C.rtlsdr_get_tuner_gain((*C.rtlsdr_dev_t)(c.dev)))
	if gain == 0 {
		err = -1
	}
	return
}

// Set the intermediate frequency gain for the device.
//
// Intermediate frequency gain stage number (1 to 6 for E4000).
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
//
// int rtlsdr_set_tuner_if_gain(rtlsdr_dev_t *dev, int stage, int gain);
// rtlsdr_set_tuner_if_gain returns 0 on success
func (c *Context) SetTunerIfGain(stage, gain int) (err int) {
	return int(C.rtlsdr_set_tuner_if_gain((*C.rtlsdr_dev_t)(c.dev),
		C.int(stage),
		C.int(gain)))
}

// Set the gain mode (automatic/manual) for the device.
// Manual gain mode must be enabled for the gain setter function to work.
//
// \param dev the device handle given by rtlsdr_open()
// \param manual gain mode, 1 means manual gain mode shall be enabled.
// \return 0 on success
// int rtlsdr_set_tuner_gain_mode(rtlsdr_dev_t *dev, int manual);
// rtlsdr_set_tuner_gain_mode returns 0 on success
func (c *Context) SetTunerGainMode(manual int) (err int) {
	return int(C.rtlsdr_set_tuner_gain_mode((*C.rtlsdr_dev_t)(c.dev),
		C.int(manual)))
}

// Selects the baseband filters according to the requested sample rate.
//
// int rtlsdr_set_sample_rate(rtlsdr_dev_t *dev, uint32_t rate);
// rtlsdr_set_sample_rate returns 0 on success
func (c *Context) SetSampleRate(rate int) (err int) {
	return int(C.rtlsdr_set_sample_rate((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rate)))
}

// Get actual sample rate the device is configured to.
//
// uint32_t rtlsdr_get_sample_rate(rtlsdr_dev_t *dev);
// rtlsdr_get_sample_rate returns 0 on error or sample rate in Hz
func (c *Context) GetSampleRate() (rate, err int) {
	rate = int(C.rtlsdr_get_sample_rate((*C.rtlsdr_dev_t)(c.dev)))
	if rate == 0 {
		err = -1
	}
	return
}

// Enable test mode that returns an 8 bit counter instead of the samples.
// The counter is generated inside the RTL2832.
//
// Test mode, 1 means enabled, 0 disabled
//
// int rtlsdr_set_testmode(rtlsdr_dev_t *dev, int on);
// rtlsdr_set_testmode returns 0 on success
func (c *Context) SetTestMode(on int) (err int) {
	return int(C.rtlsdr_set_testmode((*C.rtlsdr_dev_t)(c.dev),
		C.int(on)))
}

// Enable or disable the internal digital AGC of the RTL2832.
//
// Digital AGC mode, 1 means enabled, 0 disabled
//
// int rtlsdr_set_agc_mode(rtlsdr_dev_t *dev, int on);
// rtlsdr_set_agc_mode returns 0 on success
func (c *Context) SetAgcMode(on int) (err int) {
	return int(C.rtlsdr_set_agc_mode((*C.rtlsdr_dev_t)(c.dev),
		C.int(on)))
}

// Enable or disable the direct sampling mode. When enabled, the IF mode
// of the RTL2832 is activated, and rtlsdr_set_center_freq() will control
// the IF-frequency of the DDC, which can be used to tune from 0 to 28.8 MHz
// (xtal frequency of the RTL2832).
//
// Sampling modes, 0: disabled, 1: I-ADC input enabled, 2: Q-ADC input enabled
//
// int rtlsdr_set_direct_sampling(rtlsdr_dev_t *dev, int on);
// rtlsdr_set_direct_sampling returns 0 on success
func (c *Context) SetDirectSampling(on int) (err int) {
	return int(C.rtlsdr_set_direct_sampling((*C.rtlsdr_dev_t)(c.dev),
		C.int(on)))
}

// streaming functions

// int rtlsdr_reset_buffer(rtlsdr_dev_t *dev);
// rtlsdr_reset_buffer returns 0 on success
func (c *Context) ResetBuffer() (err int) {
	return int(C.rtlsdr_reset_buffer((*C.rtlsdr_dev_t)(c.dev)))
}

// int rtlsdr_read_sync(rtlsdr_dev_t *dev, void *buf, int len, int *n_read);
// rtlsdr_read_sync returns 0 on success
func (c *Context) ReadSync(buf []uint8, len int) (n_read int, err int) {
	err = int(C.rtlsdr_read_sync((*C.rtlsdr_dev_t)(c.dev),
		(unsafe.Pointer(&buf[0])),
		C.int(len),
		(*C.int)(unsafe.Pointer(&n_read))))
	return
}

// Read samples from the device asynchronously. This function will block until
// it is being canceled using rtlsdr_cancel_async()
//
// Optional buf_num buffer count, buf_num * buf_len = overall buffer size,
// set to 0 for default buffer count (32).
// Optional buf_len buffer length, must be multiple of 512, set to 0 for
// default buffer length (16 * 32 * 512).
//
// int rtlsdr_read_async(rtlsdr_dev_t *dev, rtlsdr_read_async_cb_t cb, void *ctx, uint32_t buf_num, uint32_t buf_len);
// rtlsdr_read_async returns 0 on success
func (c *Context) ReadAsync(f ReadAsyncCb_T, userctx *UserCtx, buf_num,
	buf_len int) (err int) {
	clientCb = f
	err = int(C.rtlsdr_read_async((*C.rtlsdr_dev_t)(c.dev),
		// (C.rtlsdr_read_async_cb_t)(*(*unsafe.Pointer)(unsafe.Pointer(&GoCallback))),
		(C.rtlsdr_read_async_cb_t)(unsafe.Pointer(&GoCallback)),
		unsafe.Pointer(userctx),
		C.uint32_t(buf_num),
		C.uint32_t(buf_len)))
	log.Printf("returning...\n")
	return
}

// Cancel all pending asynchronous operations on the device.
//
// int rtlsdr_cancel_async(rtlsdr_dev_t *dev);
// rtlsdr_cancel_async returns 0 on success
func (c *Context) CancelAsync() (err int) {
	return int(C.rtlsdr_cancel_async((*C.rtlsdr_dev_t)(c.dev)))
}
