// Copyright (c) 2012-2015 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// Package gortlsdr wraps librtlsdr, which turns your Realtek RTL2832 based
// DVB dongle into a SDR receiver.

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
#ifdef MOC_TEST
// gcc -fPIC -shared -o librtlsdr_moc.so librtlsdr_moc.c
// CC="gcc -DMOCK_TEST"  go build -o gortlsdr.a rtlsdr.go exports.go
#cgo linux LDFLAGS: -L. -lrtlsdr_moc
#else
#cgo linux LDFLAGS: -lrtlsdr
#cgo darwin LDFLAGS: -lrtlsdr
#cgo windows CFLAGS: -IC:/WINDOWS/system32
#cgo windows LDFLAGS: -lrtlsdr -LC:/WINDOWS/system32
#endif

#include <stdlib.h>
#include <rtl-sdr.h>

extern void goCallback(unsigned char *buf, uint32_t len, void *ctx);
static inline rtlsdr_read_async_cb_t get_go_cb() {
	return (rtlsdr_read_async_cb_t)goCallback;
}
*/
import "C"

// Current version.
var PackageVersion = "v2.9.13"

// ReadAsyncCbT defines a user callback function type.
type ReadAsyncCbT func([]byte)

var clientCb ReadAsyncCbT

// Context is the opened device's context.
type Context C.rtlsdr_dev_t

// UserCtx defines the second parameter of the ReadAsync method
// and is meant to be type asserted in the user's callback
// function when used. It allows the user to pass in virtually
// any object and is similar to C's void*.
//
// Examples would be a channel, a device context, a buffer, etc..
//
// A channel type assertion:  c, ok := (*userctx).(chan bool)
//
// A user context assertion:  device := (*userctx).(*rtl.Context)
type UserCtx interface{}

// CustUserCtx allows a user to specify a unique callback function
// and context with each call to ReadAsync2.
// type CustUserCtx struct {
// 	ClientCb ReadAsyncCbT
// 	Userctx  *UserCtx
// }

// HwInfo holds dongle specific information.
type HwInfo struct {
	VendorID     uint16
	ProductID    uint16
	Manufact     string
	Product      string
	Serial       string
	HaveSerial   bool
	EnableIR     bool
	RemoteWakeup bool
}

const (
	eepromSize = 256
	// maxStrSize = (max string length - 2 (header bytes)) \ 2. Where each
	// info character is followed by a null char.
	maxStrSize     = 35
	strOffsetStart = 0x09
)

// SamplingMode is the sampling mode type.
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
	EINVAL                = -22
	LIBUSB_ERROR_OTHER    = -99
)

// Note, librtlsdr's SetFreqCorrection returns an error value of
// -2 when the current ppm is the same as the requested ppm, but
// gortlsdr replaces the -2 with nil. Also, most of librtlsdr's
// functions return 0 on succes and -1 when dev is invalid but
// some return 0 when dev is invalid, go figure.
const (
	libSuccess = iota * -1
	libErrorIo
	libErrorInvalidParam
	libErrorAccess
	libErrorNoDevice
	libErrorNotFound
	libErrorBusy
	libErrorTimeout
	libErrorOverflow
	libErrorPipe
	libErrorInterrupted
	libErrorNoMem
	libErrorNotSupported
	linErrorInvalidParameter = EINVAL
	libErrorOther            = LIBUSB_ERROR_OTHER
)

// Sampling modes.
const (
	SamplingNone SamplingMode = iota
	SamplingIADC
	SamplingQADC
	SamplingUnknown
)

var libErrMap = map[int]error{
	libSuccess:           nil,
	libErrorIo:           errors.New("input/output error"),
	libErrorInvalidParam: errors.New("invalid parameter(s)"),
	libErrorAccess:       errors.New("access denied (insufficient permissions)"),
	libErrorNoDevice:     errors.New("no such device (it may have been disconnected)"),
	libErrorNotFound:     errors.New("entity not found"),
	libErrorBusy:         errors.New("resource busy"),
	libErrorTimeout:      errors.New("operation timed out"),
	libErrorOverflow:     errors.New("overflow"),
	libErrorPipe:         errors.New("pipe error"),
	libErrorInterrupted:  errors.New("system call interrupted (perhaps due to signal)"),
	libErrorNoMem:        errors.New("insufficient memory"),
	libErrorNotSupported: errors.New("operation not supported or unimplemented on this platform"),
	libErrorOther:        errors.New("invalid parameter"),
	libErrorOther:        errors.New("unknown error"),
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

// libError returns a textual error description from errno.
func libError(errno int) error {
	if err, ok := libErrMap[errno]; ok {
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
	return string(m), string(p), string(s), libError(i)
}

// GetIndexBySerial returns a device index by serial id.
func GetIndexBySerial(serial string) (index int, err error) {
	cstring := C.CString(serial)
	defer C.free(unsafe.Pointer(cstring))
	index = int(C.rtlsdr_get_index_by_serial(cstring))
	switch {
	case index >= 0:
		return
	case index == -1:
		err = errors.New("serial blank")
	case index == -2:
		err = errors.New("no devices were found")
	case index == -3:
		err = errors.New("no device found with matching name")
	default:
		err = errors.New("unknown error")
	}
	return
}

// Open returns an opened device by index.
func Open(index int) (*Context, error) {
	var dev *C.rtlsdr_dev_t
	i := int(C.rtlsdr_open((**C.rtlsdr_dev_t)(&dev),
		C.uint32_t(index)))
	return (*Context)(dev), libError(i)
}

// Close closes the device.
func (dev *Context) Close() (err error) {
	i := int(C.rtlsdr_close((*C.rtlsdr_dev_t)(dev)))
	return libError(i)
}

// configuration functions

// SetXtalFreq sets the crystal oscillator frequencies.
//
// Typically both ICs use the same clock. Changing the clock may make sense if
// you are applying an external clock to the tuner or to compensate the
// frequency (and sample rate) error caused by the original (cheap) crystal.
//
// Note, call this function only if you fully understand the implications.
func (dev *Context) SetXtalFreq(rtlFreqHz, tunerFreqHz int) (err error) {
	i := int(C.rtlsdr_set_xtal_freq((*C.rtlsdr_dev_t)(dev),
		C.uint32_t(rtlFreqHz),
		C.uint32_t(tunerFreqHz)))
	return libError(i)
}

// GetXtalFreq returns the crystal oscillator frequencies.
// Typically both ICs use the same clock.
func (dev *Context) GetXtalFreq() (rtlFreqHz, tunerFreqHz int, err error) {
	i := int(C.rtlsdr_get_xtal_freq((*C.rtlsdr_dev_t)(dev),
		(*C.uint32_t)(unsafe.Pointer(&rtlFreqHz)),
		(*C.uint32_t)(unsafe.Pointer(&tunerFreqHz))))
	return rtlFreqHz, tunerFreqHz, libError(i)
}

// GetUsbStrings returns the device information. Note, strings may be empty.
func (dev *Context) GetUsbStrings() (manufact, product, serial string, err error) {
	m := make([]byte, 257) // includes space for NULL byte
	p := make([]byte, 257)
	s := make([]byte, 257)
	i := int(C.rtlsdr_get_usb_strings((*C.rtlsdr_dev_t)(dev),
		(*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])),
		(*C.char)(unsafe.Pointer(&s[0]))))
	return string(m), string(p), string(s), libError(i)
}

// WriteEeprom writes data to the EEPROM.
func (dev *Context) WriteEeprom(data []uint8, offset uint8, leng uint16) (err error) {
	i := int(C.rtlsdr_write_eeprom((*C.rtlsdr_dev_t)(dev),
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.uint8_t(offset),
		C.uint16_t(leng)))
	switch {
	default:
		err = nil
	case i == -1:
		err = errors.New("device handle is invalid")
	case i == -2:
		err = errors.New("EEPROM size exceeded")
	case i == -3:
		err = errors.New("no EEPROM was found")
	case i < -4:
		err = errors.New("unknown error")
	}
	return
}

// ReadEeprom returns data read from the EEPROM.
func (dev *Context) ReadEeprom(data []uint8, offset uint8, leng uint16) (err error) {
	i := int(C.rtlsdr_read_eeprom((*C.rtlsdr_dev_t)(dev),
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.uint8_t(offset),
		C.uint16_t(leng)))
	switch {
	default:
		err = nil
	case i == -1:
		err = errors.New("device handle is invalid")
	case i == -2:
		err = errors.New("EEPROM size exceeded")
	case i == -3:
		err = errors.New("no EEPROM was found")
	case i < -4:
		err = errors.New("unknown error")
	}
	return
}

// SetCenterFreq sets the center frequency.
func (dev *Context) SetCenterFreq(freqHz int) (err error) {
	i := int(C.rtlsdr_set_center_freq((*C.rtlsdr_dev_t)(dev),
		C.uint32_t(freqHz)))
	return libError(i)
}

// GetCenterFreq returns the tuned frequency or zero on error.
func (dev *Context) GetCenterFreq() (freqHz int) {
	return int(C.rtlsdr_get_center_freq((*C.rtlsdr_dev_t)(dev)))
}

// SetFreqCorrection sets the frequency correction.
func (dev *Context) SetFreqCorrection(ppm int) (err error) {
	i := int(C.rtlsdr_set_freq_correction((*C.rtlsdr_dev_t)(dev),
		C.int(ppm)))
	// error code -2 means the requested PPM is the same as
	// the current PPM (dev->corr == PPM)
	if i == -2 {
		return libError(0)
	}
	return libError(i)
}

// GetFreqCorrection returns the frequency correction value.
func (dev *Context) GetFreqCorrection() (ppm int) {
	return int(C.rtlsdr_get_freq_correction((*C.rtlsdr_dev_t)(dev)))
}

// GetTunerType returns the tuner type.
func (dev *Context) GetTunerType() (tunerType string) {
	t := C.rtlsdr_get_tuner_type((*C.rtlsdr_dev_t)(dev))
	if tt, ok := tunerTypes[t]; ok {
		tunerType = tt
	} else {
		tunerType = "UNKNOWN"
	}
	return
}

// GetTunerGains returns a list of supported tuner gains.
// Values are in tenths of dB, e.g. 115 means 11.5 dB.
func (dev *Context) GetTunerGains() (gainsTenthsDb []int, err error) {
	//	count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev), nil))
	i := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(dev),
		(*C.int)(unsafe.Pointer(nil))))
	if i <= 0 {
		return gainsTenthsDb, libError(i)
	}
	buf := make([]C.int, i)
	i = int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(dev),
		(*C.int)(unsafe.Pointer(&buf[0]))))
	if i <= 0 {
		return gainsTenthsDb, libError(i)
	}
	gainsTenthsDb = make([]int, i)
	for ii := 0; ii < i; ii++ {
		gainsTenthsDb[ii] = int(buf[ii])
	}

	return gainsTenthsDb, nil
}

// SetTunerGain sets the tuner gain. Note, manual gain mode
// must be enabled for this to work. Valid gain values may be
// queried using GetTunerGains.
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (dev *Context) SetTunerGain(gainTenthsDb int) (err error) {
	i := int(C.rtlsdr_set_tuner_gain((*C.rtlsdr_dev_t)(dev),
		C.int(gainTenthsDb)))
	return libError(i)
}

// SetTunerBw sets the device bandwidth.
func (dev *Context) SetTunerBw(bwHz int) (err error) {
	i := int(C.rtlsdr_set_tuner_bandwidth((*C.rtlsdr_dev_t)(dev),
		C.uint32_t(bwHz)))
	return libError(i)
}

// Not in the rtl-sdr library yet
// GetTunerBw returns the device bandwidth setting,
// zero means automatic bandwidth.
// func (dev *Context) GetTunerBw(bwHz int) {
// 	return int(C.rtlsdr_get_tuner_bandwidth((*C.rtlsdr_dev_t)(dev)))
// }

// GetTunerGain returns the tuner gain.
//
// Gain values are in tenths of dB, e.g. 115 means 11.5 dB.
func (dev *Context) GetTunerGain() (gainTenthsDb int) {
	return int(C.rtlsdr_get_tuner_gain((*C.rtlsdr_dev_t)(dev)))
}

// SetTunerIfGain sets the intermediate frequency gain.
//
// Intermediate frequency gain stage number 1 to 6.
// Gain values are in tenths of dB, e.g. -30 means -3.0 dB.
func (dev *Context) SetTunerIfGain(stage, gainTenthsDb int) (err error) {
	i := int(C.rtlsdr_set_tuner_if_gain((*C.rtlsdr_dev_t)(dev),
		C.int(stage),
		C.int(gainTenthsDb)))
	return libError(i)
}

// SetTunerGainMode sets the gain mode (automatic/manual).
// Manual gain mode must be enabled for the gain setter function to work.
func (dev *Context) SetTunerGainMode(manualMode bool) (err error) {
	mode := 0 // automatic tuner gain
	if manualMode {
		mode = 1 // manual tuner gain
	}
	i := int(C.rtlsdr_set_tuner_gain_mode((*C.rtlsdr_dev_t)(dev),
		C.int(mode)))
	return libError(i)
}

// SetSampleRate sets the sample rate.
//
// When applicable, the baseband filters are also selected based
// on the requested sample rate.
func (dev *Context) SetSampleRate(rateHz int) (err error) {
	i := int(C.rtlsdr_set_sample_rate((*C.rtlsdr_dev_t)(dev),
		C.uint32_t(rateHz)))
	return libError(i)

}

// GetSampleRate returns the sample rate.
func (dev *Context) GetSampleRate() (rateHz int) {
	return int(C.rtlsdr_get_sample_rate((*C.rtlsdr_dev_t)(dev)))
}

// SetTestMode sets device to  test mode.
//
// Test mode returns 8 bit counters instead of samples. Note,
// the counter is generated inside the device.
func (dev *Context) SetTestMode(testMode bool) (err error) {
	mode := 0 // test mode off
	if testMode {
		mode = 1 // test mode on
	}
	i := int(C.rtlsdr_set_testmode((*C.rtlsdr_dev_t)(dev),
		C.int(mode)))
	return libError(i)
}

// SetAgcMode sets the AGC mode.
func (dev *Context) SetAgcMode(AGCMode bool) (err error) {
	mode := 0 // AGC off
	if AGCMode {
		mode = 1 // AGC on
	}
	i := int(C.rtlsdr_set_agc_mode((*C.rtlsdr_dev_t)(dev),
		C.int(mode)))
	return libError(i)
}

// SetDirectSampling sets the direct sampling mode.
//
// When enabled, the IF mode of the device is activated, and
// SetCenterFreq() will control the IF-frequency of the DDC, which
// can be used to tune from 0 to 28.8 MHz (xtal frequency of the device).
func (dev *Context) SetDirectSampling(mode SamplingMode) (err error) {
	i := int(C.rtlsdr_set_direct_sampling((*C.rtlsdr_dev_t)(dev),
		C.int(mode)))
	return libError(i)
}

// GetDirectSampling returns the state of direct sampling mode.
func (dev *Context) GetDirectSampling() (mode SamplingMode, err error) {
	i := int(C.rtlsdr_get_direct_sampling((*C.rtlsdr_dev_t)(dev)))
	switch i {
	case -1:
		err = errors.New("error getting sampling mode")
	case 0:
		mode = SamplingNone
	case 1:
		mode = SamplingIADC
	case 2:
		mode = SamplingQADC
	default:
		mode = SamplingUnknown
		err = errors.New("unknown sampling mode state")
	}
	return
}

// SetOffsetTuning sets the offset tuning mode for zero-IF tuners, which
// avoids problems caused by the DC offset of the ADCs and 1/f noise.
func (dev *Context) SetOffsetTuning(enable bool) (err error) {
	mode := 0 // offset tuning off
	if enable {
		mode = 1 // offset tuning on
	}
	i := int(C.rtlsdr_set_offset_tuning((*C.rtlsdr_dev_t)(dev), C.int(mode)))
	return libError(i)
}

// GetOffsetTuning returns the offset tuning mode.
func (dev *Context) GetOffsetTuning() (enabled bool, err error) {
	i := int(C.rtlsdr_get_offset_tuning((*C.rtlsdr_dev_t)(dev)))
	switch i {
	case -1:
		err = errors.New("error getting offset tuning mode")
	case 0:
		enabled = false
	case 1:
		enabled = true
	default:
		err = errors.New("unknown offset tuning mode state")
	}
	return
}

// streaming functions

// ResetBuffer resets the streaming buffer.
func (dev *Context) ResetBuffer() (err error) {
	i := int(C.rtlsdr_reset_buffer((*C.rtlsdr_dev_t)(dev)))
	return libError(i)
}

// ReadSync performs a synchronous read of samples and returns
// the number of samples read.
func (dev *Context) ReadSync(buf []uint8, leng int) (nRead int, err error) {
	i := int(C.rtlsdr_read_sync((*C.rtlsdr_dev_t)(dev),
		unsafe.Pointer(&buf[0]),
		C.int(leng),
		(*C.int)(unsafe.Pointer(&nRead))))
	return nRead, libError(i)
}

// Due to the restrictions imposed by the new
// "Rules for passing pointers between Go and C" at
// https://github.com/golang/proposal/blob/master/design/12416-cgo-pointers.md
// and https://github.com/golang/go/issues/12416
// https://groups.google.com/forum/#!topic/golang-dev/S7zPrUEkbKs
// https://go-review.googlesource.com/#/c/16003/
// ReadAsync no longer accepts a userdefined context parameter.

// ReadAsync reads samples asynchronously. Note, this function
// will block until canceled using CancelAsync. ReadAsyncCbT is
// a package global variable and therefore unsafe for use with
// multiple dongles.
//
// Note, please use ReadAsync2 as this method will be deprecated
// in the future.
//
// Optional bufNum buffer count, bufNum * bufLen = overall buffer size,
// set to 0 for default buffer count (32).
// Optional bufLen buffer length, must be multiple of 512, set to 0 for
// default buffer length (16 * 32 * 512).
func (dev *Context) ReadAsync(f ReadAsyncCbT, _ *UserCtx, bufNum, bufLen int) error {
	clientCb = f
	i := int(C.rtlsdr_read_async((*C.rtlsdr_dev_t)(dev),
		(C.rtlsdr_read_async_cb_t)(C.get_go_cb()),
		nil, // userctx *UserCtx
		C.uint32_t(bufNum),
		C.uint32_t(bufLen)))
	return libError(i)
}

// CancelAsync cancels all pending asynchronous operations.
func (dev *Context) CancelAsync() error {
	i := int(C.rtlsdr_cancel_async((*C.rtlsdr_dev_t)(dev)))
	return libError(i)
}

// GetHwInfo gets the dongle's information items.
func (dev *Context) GetHwInfo() (info HwInfo, err error) {
	data := make([]uint8, eepromSize)
	if err = dev.ReadEeprom(data, 0, eepromSize); err != nil {
		return
	}
	if (data[0] != 0x28) || (data[1] != 0x32) {
		err = errors.New("no valid RTL2832 EEPROM header")
		return
	}
	info.VendorID = (uint16(data[3]) << 8) | uint16(data[2])
	info.ProductID = (uint16(data[5]) << 8) | uint16(data[4])
	if data[6] == 0xA5 {
		info.HaveSerial = true
	}
	if t := data[7] & 0x01; t == 1 {
		info.RemoteWakeup = true
	}
	if t := data[7] & 0x02; t == 2 {
		info.EnableIR = true
	}
	info.Manufact, info.Product, info.Serial, err = GetStringDescriptors(data)
	return
}

// SetHwInfo sets the dongle's information items.
func (dev *Context) SetHwInfo(info HwInfo) (err error) {
	data := make([]uint8, eepromSize)
	data[0] = 0x28
	data[1] = 0x32
	data[2] = uint8(info.VendorID)
	data[3] = uint8(info.VendorID >> 8)
	data[4] = uint8(info.ProductID)
	data[5] = uint8(info.ProductID >> 8)
	if info.HaveSerial == true {
		data[6] = 0xA5
	}

	if info.RemoteWakeup == true {
		data[7] = data[7] | 0x01
	}

	if info.EnableIR == true {
		data[7] = data[7] | 0x02
	}
	if err = SetStringDescriptors(info, data); err != nil {
		return err
	}
	return dev.WriteEeprom(data, 0, eepromSize)
}

// GetStringDescriptors gets the manufacturer, product, and serial
// strings from the hardware's eeprom.
func GetStringDescriptors(data []uint8) (manufact, product, serial string, err error) {
	pos := strOffsetStart
	for _, v := range []*string{&manufact, &product, &serial} {
		l := int(data[pos])
		if l > (maxStrSize*2)+2 {
			err = errors.New("string value too long")
			return
		}
		if data[pos+1] != 0x03 {
			err = errors.New("string descriptor invalid")
			return
		}
		j := 0
		k := 0
		m := make([]uint8, l-2)
		for j = 2; j < l; j += 2 {
			m[k] = data[pos+j]
			k++
		}
		*v = string(m)
		pos += j
	}
	return
}

// SetStringDescriptors sets the manufacturer, product, and serial
// strings on the hardware's eeprom.
func SetStringDescriptors(info HwInfo, data []uint8) (err error) {
	e := ""
	if len(info.Manufact) > maxStrSize {
		e += "Manufact:"
	}
	if len(info.Product) > maxStrSize {
		e += "Product:"
	}
	if len(info.Serial) > maxStrSize {
		e += "Serial:"
	}
	if len(e) != 0 {
		err = errors.New(e + " string/s too long")
		return
	}
	pos := strOffsetStart
	for _, v := range []string{info.Manufact, info.Product, info.Serial} {
		data[pos] = uint8((len(v) * 2) + 2)
		data[pos+1] = 0x03
		i := 0
		j := 0
		for i = 2; i <= len(v)*2; i += 2 {
			data[pos+i] = v[j]
			data[pos+i+1] = 0x00
			j++
		}
		pos += i
	}
	return
}
