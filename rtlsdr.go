// Copyright (c) 2012-2016 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// Package gortlsdr wraps librtlsdr, which turns your Realtek RTL2832 based
// DVB dongle into a SDR receiver.

package rtlsdr

import (
	"bytes"
	"errors"
	"unsafe"
)

/*
// On linux, you may need to unload the kernel DVB driver via:
//     $ sudo rmmod dvb_usb_rtl28xxu rtl2832
// If building libusb from source, to regenerate the configure file use:
//     $ autoreconf -fvi

#cgo !windows LDFLAGS: -lrtlsdr
#cgo windows CFLAGS: -IC:/WINDOWS/system32
#cgo windows LDFLAGS: -lrtlsdr -LC:/WINDOWS/system32

#include <stdlib.h>
#ifdef mock
#include "rtl-sdr_moc.h"
#else
#include <rtl-sdr.h>
#endif

extern void goRTLSDRCallback(unsigned char *buf, uint32_t len, void *ctx);
static inline rtlsdr_read_async_cb_t get_go_cb() {
	return (rtlsdr_read_async_cb_t)goRTLSDRCallback;
}
*/
import "C"

// PackageVersion is the current version
var PackageVersion = "v2.10.0"

// MaxDevices is the theoretical maximum number of devices
// allowed on a single USB 2.0 controller
const MaxDevices = 127

// ReadAsyncCbT defines a user callback function type.
// FIXME add user context
// TODO add device ? many advaced rtl-sdr software use device inside callback
type ReadAsyncCbT func([]byte)
type ReadAsyncCbT2 func(*Context, []byte, interface{})

var contexts [MaxDevices]*Context

// Context is the opened device's context.
type Context struct {
	rtldev    *C.rtlsdr_dev_t
	clientCb  ReadAsyncCbT
	clientCb2 ReadAsyncCbT2
	idx       int
	userCtx   interface{}
}

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
type UserCtx interface{} // TODO remove this type, better use interface{} ?

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
	// EepromSize is the char size of the EEPROM
	EepromSize = 256
	// MaxStrSize = (max string length - 2 (header bytes)) \ 2,
	// where each info char is followed by a null char.
	MaxStrSize = 35
	// StrOffsetStart is the string descriptor offset start
	StrOffsetStart = 0x09
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
	LIBUSB_ERROR_OTHER    = -99
)

// Note, librtlsdr's SetFreqCorrection returns an error value of
// -2 when the current ppm is the same as the requested ppm, but
// gortlsdr replaces the -2 with nil. Also, most of librtlsdr's
// functions return 0 on success and -1 when dev is invalid but
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
	libErrorOther = LIBUSB_ERROR_OTHER
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
	libErrorOther:        errors.New("unknown error"),
}

// SamplingModes is a map of available sampling modes.
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
func GetDeviceCount() int {
	return int(C.rtlsdr_get_device_count())
}

// GetDeviceName returns the name of the device by index.
func GetDeviceName(index int) string {
	return C.GoString(C.rtlsdr_get_device_name(C.uint32_t(index)))
}

// GetDeviceUsbStrings returns the manufact, product, and serial
// information of a device based on index.
func GetDeviceUsbStrings(index int) (string, string, string, error) {
	m := [257]byte{} // includes space for NULL byte
	p := [257]byte{}
	s := [257]byte{}
	i := int(C.rtlsdr_get_device_usb_strings(C.uint32_t(index),
		(*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])),
		(*C.char)(unsafe.Pointer(&s[0]))))
	return string(bytes.Trim(m[:], "\x00")), string(bytes.Trim(p[:], "\x00")),
		string(bytes.Trim(s[:], "\x00")), libError(i)
}

// GetIndexBySerial returns the device index based on serial id.
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

// Open returns an opened device based on index where index < MaxDevices.
func Open(index int) (*Context, error) {
	var dev *C.rtlsdr_dev_t
	i := int(C.rtlsdr_open((**C.rtlsdr_dev_t)(&dev),
		C.uint32_t(index)))
	v := &Context{rtldev: dev, idx: index}
	contexts[index] = v
	return v, libError(i)
}

// Close closes the device.
func (dev *Context) Close() error {
	contexts[dev.idx] = nil
	dev.idx = -1
	i := int(C.rtlsdr_close(dev.rtldev)) // (*C.rtlsdr_dev_t)(dev)))
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
func (dev *Context) SetXtalFreq(rtlFreqHz, tunerFreqHz int) error {
	i := int(C.rtlsdr_set_xtal_freq(dev.rtldev,
		C.uint32_t(rtlFreqHz),
		C.uint32_t(tunerFreqHz)))
	return libError(i)
}

// GetXtalFreq returns the crystal oscillator frequencies (rtlFreqHz and
// tunerFreqHz). Typically both ICs use the same clock.
func (dev *Context) GetXtalFreq() (int, int, error) {
	var rtlFreqHz, tunerFreqHz C.uint32_t
	i := int(C.rtlsdr_get_xtal_freq(dev.rtldev,
		&rtlFreqHz,
		&tunerFreqHz))
	return int(rtlFreqHz), int(tunerFreqHz), libError(i)
}

// GetUsbStrings returns the manufact, product, and serial information
// of the device. Note, strings may be empty.
func (dev *Context) GetUsbStrings() (string, string, string, error) {
	m := [257]byte{} // includes space for NULL byte
	p := [257]byte{}
	s := [257]byte{}
	i := int(C.rtlsdr_get_usb_strings(dev.rtldev,
		(*C.char)(unsafe.Pointer(&m[0])),
		(*C.char)(unsafe.Pointer(&p[0])),
		(*C.char)(unsafe.Pointer(&s[0]))))
	return string(bytes.Trim(m[:], "\x00")), string(bytes.Trim(p[:], "\x00")),
		string(bytes.Trim(s[:], "\x00")), libError(i)
}

// WriteEeprom writes data to the EEPROM.
func (dev *Context) WriteEeprom(data []uint8, offset uint8, leng uint16) (err error) {
	i := int(C.rtlsdr_write_eeprom(dev.rtldev,
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
	i := int(C.rtlsdr_read_eeprom(dev.rtldev,
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
func (dev *Context) SetCenterFreq(freqHz int) error {
	i := int(C.rtlsdr_set_center_freq(dev.rtldev,
		C.uint32_t(freqHz)))
	return libError(i)
}

// GetCenterFreq returns the tuned frequency, or zero on error.
func (dev *Context) GetCenterFreq() int {
	return int(C.rtlsdr_get_center_freq(dev.rtldev))
}

// SetFreqCorrection sets the frequency correction.
func (dev *Context) SetFreqCorrection(ppm int) error {
	i := int(C.rtlsdr_set_freq_correction(dev.rtldev,
		C.int(ppm)))
	// error code -2 means the requested PPM is the same as
	// the current PPM (dev->corr == PPM)
	if i == -2 {
		return nil
	}
	return libError(i)
}

// GetFreqCorrection returns the frequency correction value in ppm.
func (dev *Context) GetFreqCorrection() int {
	return int(C.rtlsdr_get_freq_correction(dev.rtldev))
}

// GetTunerType returns the tuner type.
func (dev *Context) GetTunerType() (tunerType string) {
	t := C.rtlsdr_get_tuner_type(dev.rtldev)
	if tt, ok := tunerTypes[t]; ok {
		tunerType = tt
	} else {
		tunerType = "UNKNOWN"
	}
	return
}

// GetTunerGains returns a list of supported tuner gains.
//
// Values are in tenths of dB, e.g. 115 means 11.5 dB.
func (dev *Context) GetTunerGains() ([]int, error) {
	// count := int(C.rtlsdr_get_tuner_gains((*C.rtlsdr_dev_t)(c.dev), nil))
	i := int(C.rtlsdr_get_tuner_gains(dev.rtldev,
		(*C.int)(unsafe.Pointer(nil))))
	if i <= 0 {
		return nil, libError(i)
	}
	buf := make([]C.int, i)
	i = int(C.rtlsdr_get_tuner_gains(dev.rtldev,
		(*C.int)(unsafe.Pointer(&buf[0]))))
	if i <= 0 {
		return nil, libError(i)
	}
	gainsTenthsDb := make([]int, i)
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
func (dev *Context) SetTunerGain(gainTenthsDb int) error {
	i := int(C.rtlsdr_set_tuner_gain(dev.rtldev,
		C.int(gainTenthsDb)))
	return libError(i)
}

// SetTunerBw sets the device bandwidth.
func (dev *Context) SetTunerBw(bwHz int) error {
	i := int(C.rtlsdr_set_tuner_bandwidth(dev.rtldev,
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
func (dev *Context) GetTunerGain() int {
	return int(C.rtlsdr_get_tuner_gain(dev.rtldev))
}

// SetTunerIfGain sets the intermediate frequency gain.
//
// Intermediate frequency gain stage number 1 to 6.
// Gain values are in tenths of dB, e.g. -30 means -3.0 dB.
func (dev *Context) SetTunerIfGain(stage, gainTenthsDb int) error {
	i := int(C.rtlsdr_set_tuner_if_gain(dev.rtldev,
		C.int(stage),
		C.int(gainTenthsDb)))
	return libError(i)
}

// SetTunerGainMode sets the gain mode (automatic/manual).
// Manual gain mode must be enabled for the gain setter function to work.
func (dev *Context) SetTunerGainMode(manualMode bool) error {
	mode := 0 // automatic tuner gain
	if manualMode {
		mode = 1 // manual tuner gain
	}
	i := int(C.rtlsdr_set_tuner_gain_mode(dev.rtldev,
		C.int(mode)))
	return libError(i)
}

// SetSampleRate sets the sample rate.
//
// When applicable, the baseband filters are also selected based
// on the requested sample rate.
func (dev *Context) SetSampleRate(rateHz int) error {
	i := int(C.rtlsdr_set_sample_rate(dev.rtldev,
		C.uint32_t(rateHz)))
	return libError(i)

}

// GetSampleRate returns the sample rate in Hz.
func (dev *Context) GetSampleRate() int {
	return int(C.rtlsdr_get_sample_rate(dev.rtldev))
}

// SetTestMode sets device to  test mode.
//
// Test mode returns 8 bit counters instead of samples. Note,
// the counter is generated inside the device.
func (dev *Context) SetTestMode(testMode bool) error {
	mode := 0 // test mode off
	if testMode {
		mode = 1 // test mode on
	}
	i := int(C.rtlsdr_set_testmode(dev.rtldev,
		C.int(mode)))
	return libError(i)
}

// SetAgcMode sets the AGC mode.
func (dev *Context) SetAgcMode(AGCMode bool) error {
	mode := 0 // AGC off
	if AGCMode {
		mode = 1 // AGC on
	}
	i := int(C.rtlsdr_set_agc_mode(dev.rtldev,
		C.int(mode)))
	return libError(i)
}

// SetDirectSampling sets the direct sampling mode.
//
// When enabled, the IF mode of the device is activated, and
// SetCenterFreq() will control the IF-frequency of the DDC, which
// can be used to tune from 0 to 28.8 MHz (xtal frequency of the device).
func (dev *Context) SetDirectSampling(mode SamplingMode) error {
	i := int(C.rtlsdr_set_direct_sampling(dev.rtldev,
		C.int(mode)))
	return libError(i)
}

// GetDirectSampling returns the state of direct sampling mode.
func (dev *Context) GetDirectSampling() (mode SamplingMode, err error) {
	i := int(C.rtlsdr_get_direct_sampling(dev.rtldev))
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
func (dev *Context) SetOffsetTuning(enable bool) error {
	mode := 0 // offset tuning off
	if enable {
		mode = 1 // offset tuning on
	}
	i := int(C.rtlsdr_set_offset_tuning(dev.rtldev, C.int(mode)))
	return libError(i)
}

// GetOffsetTuning returns the offset tuning mode.
func (dev *Context) GetOffsetTuning() (enabled bool, err error) {
	i := int(C.rtlsdr_get_offset_tuning(dev.rtldev))
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
func (dev *Context) ResetBuffer() error {
	i := int(C.rtlsdr_reset_buffer(dev.rtldev))
	return libError(i)
}

// ReadSync performs a synchronous read of samples and returns
// the number of samples read.
func (dev *Context) ReadSync(buf []uint8, leng int) (int, error) {
	var nRead C.int
	// FIXME use len(buf) for leng ? API change
	i := int(C.rtlsdr_read_sync(dev.rtldev,
		unsafe.Pointer(&buf[0]),
		C.int(leng),
		&nRead))
	return int(nRead), libError(i)
}

// ReadAsync reads samples asynchronously. Note, this function
// will block until canceled using CancelAsync
//
// Note, please use ReadAsync2 as this method will be deprecated
// in the future.
//
// Optional bufNum buffer count, bufNum * bufLen = overall buffer size,
// set to 0 for default buffer count (32).
// Optional bufLen buffer length, must be multiple of 512, set to 0 for
// default buffer length (16 * 32 * 512).
func (dev *Context) ReadAsync(f ReadAsyncCbT, u UserCtx, bufNum, bufLen int) error {
	dev.clientCb = f
	dev.clientCb2 = nil
	dev.userCtx = u
	i := int(C.rtlsdr_read_async(dev.rtldev,
		(C.rtlsdr_read_async_cb_t)(C.get_go_cb()),
		unsafe.Pointer(uintptr(dev.idx)),
		C.uint32_t(bufNum),
		C.uint32_t(bufLen)))
	return libError(i)
}

// ReadAsync2 reads samples asynchronously. Note, this function
// will block until canceled using CancelAsync
//
// Optional bufNum buffer count, bufNum * bufLen = overall buffer size,
// set to 0 for default buffer count (32).
// Optional bufLen buffer length, must be multiple of 512, set to 0 for
// default buffer length (16 * 32 * 512).
func (dev *Context) ReadAsync2(f ReadAsyncCbT2, u UserCtx, bufNum, bufLen int) error {
	dev.clientCb2 = f
	dev.clientCb = nil
	dev.userCtx = u
	i := int(C.rtlsdr_read_async(dev.rtldev,
		(C.rtlsdr_read_async_cb_t)(C.get_go_cb()),
		unsafe.Pointer(uintptr(dev.idx)),
		C.uint32_t(bufNum),
		C.uint32_t(bufLen)))
	return libError(i)
}

// CancelAsync cancels all pending asynchronous operations.
func (dev *Context) CancelAsync() error {
	i := int(C.rtlsdr_cancel_async(dev.rtldev))
	return libError(i)
}

// GetHwInfo gets the dongle's information items.
func (dev *Context) GetHwInfo() (info HwInfo, err error) {
	data := make([]uint8, EepromSize)
	if err = dev.ReadEeprom(data, 0, EepromSize); err != nil {
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
func (dev *Context) SetHwInfo(info HwInfo) error {
	data := make([]uint8, EepromSize)
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
	if err := SetStringDescriptors(info, data); err != nil {
		return err
	}
	return dev.WriteEeprom(data, 0, EepromSize)
}

// GetStringDescriptors gets the manufacturer, product,
// and serial strings from data.
func GetStringDescriptors(data []uint8) (manufact, product, serial string, err error) {
	pos := StrOffsetStart
	for _, v := range []*string{&manufact, &product, &serial} {
		l := int(data[pos])
		if l > (MaxStrSize*2)+2 {
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
		*v = string(bytes.Trim(m, "\x00"))
		pos += j
	}
	return
}

// SetStringDescriptors sets the manufacturer, product,
// and serial strings in data.
func SetStringDescriptors(info HwInfo, data []uint8) error {
	e := ""
	if len(info.Manufact) > MaxStrSize {
		e += "Manufact:"
	}
	if len(info.Product) > MaxStrSize {
		e += "Product:"
	}
	if len(info.Serial) > MaxStrSize {
		e += "Serial:"
	}
	if len(e) != 0 {
		return errors.New(e + " string/s too long")
	}
	pos := StrOffsetStart
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
	return nil
}
