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
import "C"

import (
	//	"log"
	//	"reflect"
	"unsafe"
)

var PackageVersion string = "v0.1"

type Context struct {
	dev *C.rtlsdr_dev_t
}

const (
	/* tuner types */
	TUNER_UNKNOWN = C.RTLSDR_TUNER_UNKNOWN
	TUNER_E4000   = C.RTLSDR_TUNER_E4000
	TUNER_FC0012  = C.RTLSDR_TUNER_FC0012
	TUNER_FC0013  = C.RTLSDR_TUNER_FC0013
	TUNER_FC2580  = C.RTLSDR_TUNER_FC2580
	TUNER_R820T   = C.RTLSDR_TUNER_R820T

	/* default parameter settings */
	DEFAULT_GAIN      = "auto"
	DEFAULT_FC        = 80e6
	DEFAULT_RS        = 1.024e6
	DEFAULT_READ_SIZE = 1024
	CRYSTAL_FREQ      = 28800000

	DEFAULT_SAMPLE_RATE      = 2048000
	DEFAULT_ASYNC_BUF_NUMBER = 32
	DEFAULT_BUF_LENGTH       = (16 * 16384)
	MINIMAL_BUF_LENGTH       = 512
	MAXIMAL_BUF_LENGTH       = (256 * 16384)
)

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
func GetDeviceUsbStrings(index int) (manufact, product, serial string, err int) {
	m := make([]byte, 257) // includes space for NULL byte
	p := make([]byte, 257)
	s := make([]byte, 257)
	err = int(C.rtlsdr_get_device_usb_strings(C.uint32_t(index), (*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])), (*C.char)(unsafe.Pointer(&s[0]))))
	return string(m), string(p), string(s), err
}

// Open returns a valid device's context.
//
// int rtlsdr_open(rtlsdr_dev_t **dev, uint32_t index);
func Open(index int) (c *Context, err int) {
	c = &Context{}
	err = int(C.rtlsdr_open((**C.rtlsdr_dev_t)(&c.dev), C.uint32_t(index)))
	return
}

// Close closes a previously opened device context.
//
// int rtlsdr_close(rtlsdr_dev_t *dev);
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
func (c *Context) SetXtalFreq(rtl_freq, tuner_freq int) (err int) {
	return int(C.rtlsdr_set_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rtl_freq), C.uint32_t(tuner_freq)))
}

// Get crystal oscillator frequencies used for the RTL2832 and the tuner IC.
//
// Usually both ICs use the same clock.
// Frequency values are in Hz.
//
// int rtlsdr_get_xtal_freq(rtlsdr_dev_t *dev, uint32_t *rtl_freq, uint32_t *tuner_freq);
func (c *Context) GetXtalFreq() (rtl_freq, tuner_freq, err int) {
	err = int(C.rtlsdr_get_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint32_t)(unsafe.Pointer(&rtl_freq)),
		(*C.uint32_t)(unsafe.Pointer(&tuner_freq))))
	return
}

// Get USB device strings.
//
// int rtlsdr_get_usb_strings(rtlsdr_dev_t *dev, char *manufact, char *product, char *serial);
func (c *Context) GetUsbStrings() (manufact, product, serial string, err int) {
	m := make([]byte, 257) // includes space for NULL byte
	p := make([]byte, 257)
	s := make([]byte, 257)
	err = int(C.rtlsdr_get_usb_strings((*C.rtlsdr_dev_t)(c.dev), (*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])), (*C.char)(unsafe.Pointer(&s[0]))))
	return string(m), string(p), string(s), err
}

// Set the device center frequency.
//
// int rtlsdr_set_center_freq(rtlsdr_dev_t *dev, uint32_t freq);
func (c *Context) SetCenterFreq(freq int) (err int) {
	return int(C.rtlsdr_set_center_freq((*C.rtlsdr_dev_t)(c.dev), C.uint32_t(freq)))
}

// Get actual frequency the device is tuned to.
//
// Frequency values are in Hz.
//
// uint32_t rtlsdr_get_center_freq(rtlsdr_dev_t *dev);
func (c *Context) GetCenterFreq() (freq int) {
	return int(C.rtlsdr_get_center_freq((*C.rtlsdr_dev_t)(c.dev)))
}

// Set the frequency correction value for the device.
//
// Frequency values are in Hz.
//
// int rtlsdr_set_freq_correction(rtlsdr_dev_t *dev, int ppm);
func (c *Context) SetFreqCorrection(ppm int) (err int) {
	return int(C.rtlsdr_set_freq_correction((*C.rtlsdr_dev_t)(c.dev), C.int(ppm)))
}

// Get actual frequency correction value of the device.
//
// Correction value in ppm (parts per million).
//
// int rtlsdr_get_freq_correction(rtlsdr_dev_t *dev);
func (c *Context) GetFreqCorrection() (freq int) {
	return int(C.rtlsdr_get_freq_correction((*C.rtlsdr_dev_t)(c.dev)))
}

// Get the tuner type.
//
// enum rtlsdr_tuner rtlsdr_get_tuner_type(rtlsdr_dev_t *dev);
func (c *Context) GetTunerType() (rtlsdr_tuner int) {
	return int(C.rtlsdr_get_tuner_type((*C.rtlsdr_dev_t)(c.dev)))
}

// Get a list of gains supported by the tuner.
//
// A NULL second parameter requests the gain count.
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
//
// int rtlsdr_get_tuner_gains(rtlsdr_dev_t *dev, int *gains);
func (c *Context) GetTunerGains() (gains []int) {
	//	count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev), nil))
	count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev),
		(*C.int)(unsafe.Pointer(uintptr(0)))))
	gains = make([]int, count)
	if count != 0 {
		C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev), (*C.int)(unsafe.Pointer(&gains[0])))
	}
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
func (c *Context) SetTunerGain(gain int) (err int) {
	return int(C.rtlsdr_set_tuner_gain((*C.rtlsdr_dev_t)(c.dev), C.int(gain)))
}

// Get actual gain the device is configured to.
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
//
// int rtlsdr_get_tuner_gain(rtlsdr_dev_t *dev);
func (c *Context) GetTunerGain() (gain int) {
	return int(C.rtlsdr_get_tuner_gain((*C.rtlsdr_dev_t)(c.dev)))
}

// Set the intermediate frequency gain for the device.
//
// Intermediate frequency gain stage number (1 to 6 for E4000).
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
//
// int rtlsdr_set_tuner_if_gain(rtlsdr_dev_t *dev, int stage, int gain);
func (c *Context) SetTunerIfGain(stage, gain int) (err int) {
	return int(C.rtlsdr_set_tuner_if_gain((*C.rtlsdr_dev_t)(c.dev),
		C.int(stage), C.int(gain)))
}

// Set the gain mode (automatic/manual) for the device.
// Manual gain mode must be enabled for the gain setter function to work.
//
// \param dev the device handle given by rtlsdr_open()
// \param manual gain mode, 1 means manual gain mode shall be enabled.
// \return 0 on success
// int rtlsdr_set_tuner_gain_mode(rtlsdr_dev_t *dev, int manual);
func (c *Context) SetTunerGainMode(manual int) (err int) {
	return int(C.rtlsdr_set_tuner_gain_mode((*C.rtlsdr_dev_t)(c.dev),
		C.int(manual)))
}

// Selects the baseband filters according to the requested sample rate.
//
// int rtlsdr_set_sample_rate(rtlsdr_dev_t *dev, uint32_t rate);
func (c *Context) SetSampleRate(rate int) (err int) {
	return int(C.rtlsdr_set_sample_rate((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rate)))
}

// Get actual sample rate the device is configured to.
//
// uint32_t rtlsdr_get_sample_rate(rtlsdr_dev_t *dev);
func (c *Context) GetSampleRate() (rate int) {
	return int(C.rtlsdr_get_sample_rate((*C.rtlsdr_dev_t)(c.dev)))
}

// Enable test mode that returns an 8 bit counter instead of the samples.
// The counter is generated inside the RTL2832.
//
// Test mode, 1 means enabled, 0 disabled
//
// int rtlsdr_set_testmode(rtlsdr_dev_t *dev, int on);
func (c *Context) SetTestMode(on int) (err int) {
	return int(C.rtlsdr_set_testmode((*C.rtlsdr_dev_t)(c.dev),
		C.int(on)))
}

// Enable or disable the internal digital AGC of the RTL2832.
//
// Digital AGC mode, 1 means enabled, 0 disabled
//
// int rtlsdr_set_agc_mode(rtlsdr_dev_t *dev, int on);
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
func (c *Context) SetDirectSampling(on int) (err int) {
	return int(C.rtlsdr_set_direct_sampling((*C.rtlsdr_dev_t)(c.dev),
		C.int(on)))
}

// streaming functions

// int rtlsdr_reset_buffer(rtlsdr_dev_t *dev);
func (c *Context) ResetBuffer() (err int) {
	return int(C.rtlsdr_reset_buffer((*C.rtlsdr_dev_t)(c.dev)))
}

// int rtlsdr_read_sync(rtlsdr_dev_t *dev, void *buf, int len, int *n_read);
func (c *Context) ReadSync(buf []uint8, len int) (n_read int, err int) {
	err = int(C.rtlsdr_read_sync((*C.rtlsdr_dev_t)(c.dev), (unsafe.Pointer(&buf[0])),
		C.int(len), (*C.int)(unsafe.Pointer(&n_read))))
	return
}

// typedef void(*rtlsdr_read_async_cb_t)(unsigned char *buf, uint32_t len, void *ctx);
type ReadAsyncCb_T func(buf []uint8, len uint32, userdata interface{})

// Read samples from the device asynchronously. This function will block until
// it is being canceled using rtlsdr_cancel_async()
//
// Optional buf_num buffer count, buf_num * buf_len = overall buffer size,
// set to 0 for default buffer count (32).
// Optional buf_len buffer length, must be multiple of 512, set to 0 for
// default buffer length (16 * 32 * 512).
//
// int rtlsdr_read_async(rtlsdr_dev_t *dev, rtlsdr_read_async_cb_t cb, void *ctx, uint32_t buf_num, uint32_t buf_len);
func (c *Context) ReadAsync(f ReadAsyncCb_T, userdata *interface{}, buf_num, buf_len int) (n_read int, err int) {
	err = int(C.rtlsdr_read_async((*C.rtlsdr_dev_t)(c.dev),
		(C.rtlsdr_read_async_cb_t)(unsafe.Pointer(&f)),
		unsafe.Pointer(userdata),
		C.uint32_t(buf_num),
		C.uint32_t(buf_len)))
	return
}

// Cancel all pending asynchronous operations on the device.
//
// int rtlsdr_cancel_async(rtlsdr_dev_t *dev);
func (c *Context) CancelAsync() (err int) {
	return int(C.rtlsdr_cancel_async((*C.rtlsdr_dev_t)(c.dev)))
}
