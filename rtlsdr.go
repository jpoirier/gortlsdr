// Copyright (c) 2012 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// Package gortlsdr wraps librtlsdr, which turns your Realtek RTL2832 based
// DVB dongle into a SDR receiver.
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

extern void go_callback(unsigned char *buf, uint32_t len, void *ctx);
rtlsdr_read_async_cb_t get_go_cb() {
	return (rtlsdr_read_async_cb_t)go_callback;
}
*/
import "C"

var PackageVersion string = "v2.2"

type Context struct {
	dev *C.rtlsdr_dev_t
}

type UserCtx interface{}

type SamplingMode int

const (
	// default parameter settings
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
}

type ReadAsyncCb_T func([]byte, *UserCtx)

var clientCb ReadAsyncCb_T

// given error int, return corresponding Go error
func libusbError(errno int) error {
	if err, ok := errMap[errno]; ok {
		return err
	} else {
		return errors.New("unknown error")
	}
}

// GetDeviceCount gets the number of valid USB dongles detected.
func GetDeviceCount() (count int) {
	return int(C.rtlsdr_get_device_count())
}

// GetDeviceName gets the name of the USB dongle device via index,
// e.g. from an index returned from calling GainValues.
func GetDeviceName(index int) (name string) {
	return C.GoString(C.rtlsdr_get_device_name(C.uint32_t(index)))
}

// Get USB device strings.
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

// Get device index by USB serial string descriptor.
// Returns device index of first device where the name matched.
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

// Open returns a valid device's context.
func Open(index int) (c *Context, err error) {
	c = &Context{}
	i := int(C.rtlsdr_open((**C.rtlsdr_dev_t)(&c.dev),
		C.uint32_t(index)))
	return c, libusbError(i)
}

// Close closes a previously opened device context.
func (c *Context) Close() (err error) {
	i := int(C.rtlsdr_close((*C.rtlsdr_dev_t)(c.dev)))
	return libusbError(i)
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
func (c *Context) SetXtalFreq(rtl_freq, tuner_freq int) (err error) {
	i := int(C.rtlsdr_set_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rtl_freq),
		C.uint32_t(tuner_freq)))
	return libusbError(i)
}

// Get crystal oscillator frequencies used for the RTL2832 and the tuner IC.
//
// Usually both ICs use the same clock.
// Frequency values are in Hz.
func (c *Context) GetXtalFreq() (rtl_freq, tuner_freq int, err error) {
	i := int(C.rtlsdr_get_xtal_freq((*C.rtlsdr_dev_t)(c.dev),
		(*C.uint32_t)(unsafe.Pointer(&rtl_freq)),
		(*C.uint32_t)(unsafe.Pointer(&tuner_freq))))
	return rtl_freq, tuner_freq, libusbError(i)
}

// Get USB device strings.
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

// Write the device EEPROM
//
// data = buffer of data to be written, offset = address where the data should be written,
// leng = length of the data
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

// Read the device EEPROM
//
// data = buffer where the data should be written, offset = address where the data should be read from,
// leng = length of the data
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

// Set the device center frequency.
//
// Frequency values are in Hz.
func (c *Context) SetCenterFreq(freq int) (err error) {
	i := int(C.rtlsdr_set_center_freq((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(freq)))
	return libusbError(i)
}

// Get actual frequency the device is tuned to.
//
// Frequency values are in Hz.
func (c *Context) GetCenterFreq() (freq int) {
	return int(C.rtlsdr_get_center_freq((*C.rtlsdr_dev_t)(c.dev)))
}

// Set the frequency correction value for the device.
//
// Frequency values are in Hz.
func (c *Context) SetFreqCorrection(freq int) (err error) {
	i := int(C.rtlsdr_set_freq_correction((*C.rtlsdr_dev_t)(c.dev),
		C.int(freq)))
	return libusbError(i)
}

// Get actual frequency correction value of the device.
//
// Correction value in ppm (parts per million).
func (c *Context) GetFreqCorrection() (ppm int) {
	return int(C.rtlsdr_get_freq_correction((*C.rtlsdr_dev_t)(c.dev)))
}

// Get the tuner type.
func (c *Context) GetTunerType() (tunerType string) {
	t := C.rtlsdr_get_tuner_type((*C.rtlsdr_dev_t)(c.dev))
	if tt, ok := tunerTypes[t]; ok {
		tunerType = tt
	} else {
		tunerType = "UNKNOWN"
	}
	return
}

// Get a list of gains supported by the tuner.
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (c *Context) GetTunerGains() (gains []int, err error) {
	//	count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev), nil))
	i := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev),
		(*C.int)(unsafe.Pointer(uintptr(0)))))
	if i < 0 {
		return gains, libusbError(i)
	} else if i == 0 {
		return gains, nil
	}
	gains = make([]int, i)
	C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev),
		(*C.int)(unsafe.Pointer(&gains[0])))
	return gains, nil
}

// Set the gain for the device.
// Manual gain mode must be enabled for this to work.
//
// Valid gain values (in tenths of a dB) for the E4000 tuner:
// -10, 15, 40, 65, 90, 115, 140, 165, 190, 215, 240, 290,
// 340, 420, 430, 450, 470, 490
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (c *Context) SetTunerGain(gain int) (err error) {
	i := int(C.rtlsdr_set_tuner_gain((*C.rtlsdr_dev_t)(c.dev),
		C.int(gain)))
	return libusbError(i)
}

// Get actual gain the device is configured to.
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (c *Context) GetTunerGain() (gain int) {
	return int(C.rtlsdr_get_tuner_gain((*C.rtlsdr_dev_t)(c.dev)))
}

// Set the intermediate frequency gain for the device.
//
// Intermediate frequency gain stage number (1 to 6 for E4000).
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (c *Context) SetTunerIfGain(stage, gain int) (err error) {
	i := int(C.rtlsdr_set_tuner_if_gain((*C.rtlsdr_dev_t)(c.dev),
		C.int(stage),
		C.int(gain)))
	return libusbError(i)
}

// Set the gain mode (automatic/manual) for the device.
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

// Selects the baseband filters according to the requested sample rate.
//
// Samplerate is in Hz.
func (c *Context) SetSampleRate(rate int) (err error) {
	i := int(C.rtlsdr_set_sample_rate((*C.rtlsdr_dev_t)(c.dev),
		C.uint32_t(rate)))
	return libusbError(i)

}

// Get actual sample rate the device is configured to.
//
// Samplerate is in Hz.
func (c *Context) GetSampleRate() (rate int) {
	return int(C.rtlsdr_get_sample_rate((*C.rtlsdr_dev_t)(c.dev)))
}

// Enable test mode that returns an 8 bit counter instead of the samples.
// The counter is generated inside the RTL2832.
func (c *Context) SetTestMode(testMode bool) (err error) {
	mode := 0
	if testMode {
		mode = 1
	}
	i := int(C.rtlsdr_set_testmode((*C.rtlsdr_dev_t)(c.dev),
		C.int(mode)))
	return libusbError(i)
}

// Enable or disable the internal digital AGC of the RTL2832.
func (c *Context) SetAgcMode(AGCMode bool) (err error) {
	mode := 0
	if AGCMode {
		mode = 1
	}
	i := int(C.rtlsdr_set_agc_mode((*C.rtlsdr_dev_t)(c.dev),
		C.int(mode)))
	return libusbError(i)
}

// Enable or disable the direct sampling mode. When enabled, the IF mode
// of the RTL2832 is activated, and rtlsdr_set_center_freq() will control
// the IF-frequency of the DDC, which can be used to tune from 0 to 28.8 MHz
// (xtal frequency of the RTL2832).
func (c *Context) SetDirectSampling(mode SamplingMode) (err error) {
	i := int(C.rtlsdr_set_direct_sampling((*C.rtlsdr_dev_t)(c.dev),
		C.int(mode)))
	return libusbError(i)
}

// Get state of the direct sampling mode
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

// Enable or disable offset tuning for zero-IF tuners, which allows to avoid
// problems caused by the DC offset of the ADCs and 1/f noise.
func (c *Context) SetOffsetTuning(enable bool) (err error) {
	mode := 1
	if !enable {
		mode = 0
	}
	i := int(C.rtlsdr_set_offset_tuning((*C.rtlsdr_dev_t)(c.dev), C.int(mode)))
	return libusbError(i)
}

// Get state of the offset tuning mode
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

func (c *Context) ResetBuffer() (err error) {
	i := int(C.rtlsdr_reset_buffer((*C.rtlsdr_dev_t)(c.dev)))
	return libusbError(i)
}

func (c *Context) ReadSync(buf []uint8, leng int) (n_read int, err error) {
	i := int(C.rtlsdr_read_sync((*C.rtlsdr_dev_t)(c.dev),
		unsafe.Pointer(&buf[0]),
		C.int(leng),
		(*C.int)(unsafe.Pointer(&n_read))))
	return n_read, libusbError(i)
}

// Read samples from the device asynchronously. This function will block until
// it is being canceled using CancelAsync
//
// Optional buf_num buffer count, buf_num * buf_len = overall buffer size,
// set to 0 for default buffer count (32).
// Optional buf_len buffer length, must be multiple of 512, set to 0 for
// default buffer length (16 * 32 * 512).
func (c *Context) ReadAsync(f ReadAsyncCb_T, userctx *UserCtx, buf_num,
	buf_len int) (err error) {
	clientCb = f
	i := int(C.rtlsdr_read_async((*C.rtlsdr_dev_t)(c.dev),
		(C.rtlsdr_read_async_cb_t)(C.get_go_cb()),
		unsafe.Pointer(userctx),
		C.uint32_t(buf_num),
		C.uint32_t(buf_len)))
	return libusbError(i)
}

// Cancel all pending asynchronous operations on the device.
func (c *Context) CancelAsync() (err error) {
	i := int(C.rtlsdr_cancel_async((*C.rtlsdr_dev_t)(c.dev)))
	return libusbError(i)
}
