// Copyright (c) 2012 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// Package gortlsdr wraps librtlsdr, which turns your Realtek RTL2832 based
// DVB dongle into a SDR receiver.
//
package rtlsdr

import (
	"unsafe"
	// "reflect"
)

/*
// On linux, you may need to unload the kernel DVB driver via:
//     $ sudo rmmod dvb_usb_rtl28xxu rtl2832
// If building libusb from source, to regenerate the configure file use:
//     $ autoreconf -fvi

#cgo linux LDFLAGS: -lrtlsdr -lusb
#cgo darwin LDFLAGS: -lrtlsdr
#cgo windows CFLAGS: -IC:/WINDOWS/system32
#cgo windows LDFLAGS: -lrtlsdr -LC:/WINDOWS/system32

#include <stdlib.h>
#include <rtl-sdr.h>

extern void go_callback(unsigned char *buf, uint32_t len, void *ctx);
rtlsdr_read_async_cb_t get_go_cb() {
	return (rtlsdr_read_async_cb_t)go_callback;
}
*/
import "C"

var PackageVersion string = "v1.6"

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
	Fail                  = -1

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

	GainAuto   = 0
	GainManual = 1
)

var Status = map[int]string{
	Success: "Successful",
	Fail:    "Failed",
}

var TunerType = map[int]string{
	TunerUnknown: "RTLSDR_TUNER_UNKNOWN",
	TunerE4000:   "RTLSDR_TUNER_E4000",
	TunerFC0012:  "RTLSDR_TUNER_FC0012",
	TunerFC0013:  "RTLSDR_TUNER_FC0013",
	TunerFC2580:  "RTLSDR_TUNER_FC2580",
	TunerR820T:   "RTLSDR_TUNER_R820T",
}

type ReadAsyncCb_T func([]byte, *UserCtx)

var clientCb ReadAsyncCb_T

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

// Get device index by USB serial string descriptor.
//

// returns device index of first device where the name matched
// -1 if name is NULL
// -2 if no devices were found at all
// -3 if devices were found, but none with matching name
//
// int rtlsdr_get_index_by_serial(const char *serial);
func GetIndexBySerial(serial string) (index int) {
	cstring := C.CString(serial)
	defer C.free(unsafe.Pointer(cstring))
	return int(C.rtlsdr_get_index_by_serial(cstring))
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

// Write the device EEPROM
//
// data buffer of data to be written
// offset address where the data should be written
// len length of the data
//
// int rtlsdr_write_eeprom(rtlsdr_dev_t *dev, uint8_t *data, uint8_t offset, uint16_t len);
// rtlsdr_write_eeprom returns 0 on success, -1 if device handle is invalid,
// -2 if EEPROM size is exceeded, -3 if no EEPROM was found
func (c *Context) WriteEeprom(data []uint8, offset uint8, len uint16) (err int) {
	return int(C.rtlsdr_write_eeprom((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.uint8_t(offset),
		C.uint16_t(len)))
}

// Read the device EEPROM
//
// data buffer where the data should be written
// offset address where the data should be read from
// len length of the data
//
// int rtlsdr_read_eeprom(rtlsdr_dev_t *dev, uint8_t *data, uint8_t offset, uint16_t len);
// rtlsdr_read_eeprom returns 0 on success, -1 if device handle is invalid,
// -2 if EEPROM size is exceeded, -3 if no EEPROM was found
func (c *Context) ReadEeprom(data []uint8, offset uint8, len uint16) (err int) {
	return int(C.rtlsdr_read_eeprom((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.uint8_t(offset),
		C.uint16_t(len)))
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
func (c *Context) GetCenterFreq() (freq int) {
	return int(C.rtlsdr_get_center_freq((*C.rtlsdr_dev_t)(c.dev)))
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
func (c *Context) GetFreqCorrection() (freq int) {
	return int(C.rtlsdr_get_freq_correction((*C.rtlsdr_dev_t)(c.dev)))
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
func (c *Context) GetTunerGain() (gain int) {
	return int(C.rtlsdr_get_tuner_gain((*C.rtlsdr_dev_t)(c.dev)))
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
// manual gain mode, 1 means manual gain mode shall be enabled.
//
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
func (c *Context) GetSampleRate() (rate int) {
	return int(C.rtlsdr_get_sample_rate((*C.rtlsdr_dev_t)(c.dev)))
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

// Get state of the direct sampling mode
//
// int rtlsdr_get_direct_sampling(rtlsdr_dev_t *dev);
// rtlsdr_get_direct_sampling returns -1 on error, 0 means disabled,
// 1 I-ADC input enabled, 2 Q-ADC input enabled
func (c *Context) GetDirectSampling() (err int) {
	return int(C.rtlsdr_get_direct_sampling((*C.rtlsdr_dev_t)(c.dev)))
}

// Enable or disable offset tuning for zero-IF tuners, which allows to avoid
// problems caused by the DC offset of the ADCs and 1/f noise.
//
// 0 means disabled, 1 enabled
//
// int rtlsdr_set_offset_tuning(rtlsdr_dev_t *dev, int on);
// return 0 on success
func (c *Context) SetOffsetTuning(on int) (err int) {
	return int(C.rtlsdr_set_offset_tuning((*C.rtlsdr_dev_t)(c.dev), C.int(on)))
}

// Get state of the offset tuning mode
//
// int rtlsdr_get_offset_tuning(rtlsdr_dev_t *dev);
// rtlsdr_get_offset_tuning returns -1 on error, 0 means disabled, 1 enabled
func (c *Context) GetOffsetTuning() (err int) {
	return int(C.rtlsdr_get_offset_tuning((*C.rtlsdr_dev_t)(c.dev)))
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
		unsafe.Pointer(&buf[0]),
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
		(C.rtlsdr_read_async_cb_t)(C.get_go_cb()),
		unsafe.Pointer(userctx),
		C.uint32_t(buf_num),
		C.uint32_t(buf_len)))
	return
}

// Cancel all pending asynchronous operations on the device.
//
// int rtlsdr_cancel_async(rtlsdr_dev_t *dev);
// rtlsdr_cancel_async returns 0 on success
func (c *Context) CancelAsync() (err int) {
	return int(C.rtlsdr_cancel_async((*C.rtlsdr_dev_t)(c.dev)))
}
