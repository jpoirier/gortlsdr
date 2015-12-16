/*
 * rtl-sdr, turns your Realtek RTL2832 based DVB dongle into a SDR receiver
 * Copyright (C) 2012-2013 by Steve Markgraf <steve@steve-m.de>
 * Copyright (C) 2012 by Dimitri Stolnikov <horiz0n@gmx.net>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
#ifndef __RTL_SDR_H
#define __RTL_SDR_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct rtlsdr_dev rtlsdr_dev_t;

// returns: 0 ... n
extern uint32_t rtlsdr_get_device_count(void);

// returns "" or "some name"
extern const char* rtlsdr_get_device_name(uint32_t index);


// enum libusb_error {
// 	/** Success (no error) */
// 	LIBUSB_SUCCESS = 0,

// 	/** Input/output error */
// 	LIBUSB_ERROR_IO = -1,

// 	/** Invalid parameter */
// 	LIBUSB_ERROR_INVALID_PARAM = -2,

// 	/** Access denied (insufficient permissions) */
// 	LIBUSB_ERROR_ACCESS = -3,

// 	/** No such device (it may have been disconnected) */
// 	LIBUSB_ERROR_NO_DEVICE = -4,

// 	/** Entity not found */
// 	LIBUSB_ERROR_NOT_FOUND = -5,

// 	/** Resource busy */
// 	LIBUSB_ERROR_BUSY = -6,

// 	/** Operation timed out */
// 	LIBUSB_ERROR_TIMEOUT = -7,

// 	* Overflow
// 	LIBUSB_ERROR_OVERFLOW = -8,

// 	/** Pipe error */
// 	LIBUSB_ERROR_PIPE = -9,

// 	/** System call interrupted (perhaps due to signal) */
// 	LIBUSB_ERROR_INTERRUPTED = -10,

// 	/** Insufficient memory */
// 	LIBUSB_ERROR_NO_MEM = -11,

// 	/** Operation not supported or unimplemented on this platform */
// 	LIBUSB_ERROR_NOT_SUPPORTED = -12,

// 	/* NB: Remember to update LIBUSB_ERROR_COUNT below as well as the
// 	   message strings in strerror.c when adding new error codes here. */

// 	/** Other error */
// 	LIBUSB_ERROR_OTHER = -99,
// };
// returns -2, LIBUSB_ERROR_NO_MEM, LIBUSB_ERROR_ACCESS,
// LIBUSB_ERROR_NO_DEVICE, LIBUSB_ERROR
/*!
 * Get USB device strings.
 *
 * NOTE: The string arguments must provide space for up to 256 bytes.
 *
 * \param index the device index
 * \param manufact manufacturer name, may be NULL
 * \param product product name, may be NULL
 * \param serial serial number, may be NULL
 * \return 0 on success
 */
extern int rtlsdr_get_device_usb_strings(uint32_t index,
					     char *manufact,
					     char *product,
					     char *serial);

/*!
 * Get device index by USB serial string descriptor.
 *
 * \param serial serial string of the device
 * \return device index of first device where the name matched
 * \return -1 if name is NULL
 * \return -2 if no devices were found at all
 * \return -3 if devices were found, but none with matching name
 */
extern int rtlsdr_get_index_by_serial(const char *serial);

// returns -1 no device, some libusb_error value, also write_regs
// returns 0 or -1
extern int rtlsdr_open(rtlsdr_dev_t **dev, uint32_t index);

// returns -1 no device or 0
extern int rtlsdr_close(rtlsdr_dev_t *dev);

/* configuration functions */

/*!
 * Set crystal oscillator frequencies used for the RTL2832 and the tuner IC.
 *
 * Usually both ICs use the same clock. Changing the clock may make sense if
 * you are applying an external clock to the tuner or to compensate the
 * frequency (and samplerate) error caused by the original (cheap) crystal.
 *
 * NOTE: Call this function only if you fully understand the implications.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param rtl_freq frequency value used to clock the RTL2832 in Hz
 * \param tuner_freq frequency value used to clock the tuner IC in Hz
 * \return 0 on success
 */
extern int rtlsdr_set_xtal_freq(rtlsdr_dev_t *dev, uint32_t rtl_freq,
				    uint32_t tuner_freq);

/*!
 * Get crystal oscillator frequencies used for the RTL2832 and the tuner IC.
 *
 * Usually both ICs use the same clock.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param rtl_freq frequency value used to clock the RTL2832 in Hz
 * \param tuner_freq frequency value used to clock the tuner IC in Hz
 * \return 0 on success
 */
extern int rtlsdr_get_xtal_freq(rtlsdr_dev_t *dev, uint32_t *rtl_freq,
				    uint32_t *tuner_freq);

/*!
 * Get USB device strings.
 *
 * NOTE: The string arguments must provide space for up to 256 bytes.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param manufact manufacturer name, may be NULL
 * \param product product name, may be NULL
 * \param serial serial number, may be NULL
 * \return 0 on success
 */
extern int rtlsdr_get_usb_strings(rtlsdr_dev_t *dev, char *manufact,
				      char *product, char *serial);

/*!
 * Write the device EEPROM
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param data buffer of data to be written
 * \param offset address where the data should be written
 * \param len length of the data
 * \return 0 on success
 * \return -1 if device handle is invalid
 * \return -2 if EEPROM size is exceeded
 * \return -3 if no EEPROM was found
 */
extern int rtlsdr_write_eeprom(rtlsdr_dev_t *dev, uint8_t *data,
				  uint8_t offset, uint16_t len);

/*!
 * Read the device EEPROM
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param data buffer where the data should be written
 * \param offset address where the data should be read from
 * \param len length of the data
 * \return 0 on success
 * \return -1 if device handle is invalid
 * \return -2 if EEPROM size is exceeded
 * \return -3 if no EEPROM was found
 */
extern int rtlsdr_read_eeprom(rtlsdr_dev_t *dev, uint8_t *data,
				  uint8_t offset, uint16_t len);

extern int rtlsdr_set_center_freq(rtlsdr_dev_t *dev, uint32_t freq);

/*!
 * Get actual frequency the device is tuned to.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return 0 on error, frequency in Hz otherwise
 */
extern uint32_t rtlsdr_get_center_freq(rtlsdr_dev_t *dev);

/*!
 * Set the frequency correction value for the device.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param ppm correction value in parts per million (ppm)
 * \return 0 on success
 */
extern int rtlsdr_set_freq_correction(rtlsdr_dev_t *dev, int ppm);

/*!
 * Get actual frequency correction value of the device.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return correction value in parts per million (ppm)
 */
extern int rtlsdr_get_freq_correction(rtlsdr_dev_t *dev);

// XXX: for enum returns there should be a NO_DEVICE = -1
enum rtlsdr_tuner {
	RTLSDR_TUNER_UNKNOWN = 0,
	RTLSDR_TUNER_E4000,
	RTLSDR_TUNER_FC0012,
	RTLSDR_TUNER_FC0013,
	RTLSDR_TUNER_FC2580,
	RTLSDR_TUNER_R820T,
	RTLSDR_TUNER_R828D
};

/*!
 * Get the tuner type.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return RTLSDR_TUNER_UNKNOWN on error, tuner type otherwise
 */
extern enum rtlsdr_tuner rtlsdr_get_tuner_type(rtlsdr_dev_t *dev);

/*!
 * Get a list of gains supported by the tuner.
 *
 * NOTE: The gains argument must be preallocated by the caller. If NULL is
 * being given instead, the number of available gain values will be returned.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param gains array of gain values. In tenths of a dB, 115 means 11.5 dB.
 * \return <= 0 on error, number of available (returned) gain values otherwise
 */
extern int rtlsdr_get_tuner_gains(rtlsdr_dev_t *dev, int *gains);

/*!
 * Set the gain for the device.
 * Manual gain mode must be enabled for this to work.
 *
 * Valid gain values (in tenths of a dB) for the E4000 tuner:
 * -10, 15, 40, 65, 90, 115, 140, 165, 190,
 * 215, 240, 290, 340, 420, 430, 450, 470, 490
 *
 * Valid gain values may be queried with \ref rtlsdr_get_tuner_gains function.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param gain in tenths of a dB, 115 means 11.5 dB.
 * \return 0 on success
 */
extern int rtlsdr_set_tuner_gain(rtlsdr_dev_t *dev, int gain);

/*!
 * Set the bandwidth for the device.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param bw bandwidth in Hz. Zero means automatic BW selection.
 * \return 0 on success
 */
extern int rtlsdr_set_tuner_bandwidth(rtlsdr_dev_t *dev, uint32_t bw);

/*!
 * Get actual gain the device is configured to.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return 0 on error, gain in tenths of a dB, 115 means 11.5 dB.
 */
extern int rtlsdr_get_tuner_gain(rtlsdr_dev_t *dev);

/*!
 * Set the intermediate frequency gain for the device.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param stage intermediate frequency gain stage number (1 to 6 for E4000)
 * \param gain in tenths of a dB, -30 means -3.0 dB.
 * \return 0 on success
 */
extern int rtlsdr_set_tuner_if_gain(rtlsdr_dev_t *dev, int stage, int gain);

/*!
 * Set the gain mode (automatic/manual) for the device.
 * Manual gain mode must be enabled for the gain setter function to work.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param manual gain mode, 1 means manual gain mode shall be enabled.
 * \return 0 on success
 */
extern int rtlsdr_set_tuner_gain_mode(rtlsdr_dev_t *dev, int manual);

/*!
 * Set the sample rate for the device, also selects the baseband filters
 * according to the requested sample rate for tuners where this is possible.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param samp_rate the sample rate to be set, possible values are:
 * 		    225001 - 300000 Hz
 * 		    900001 - 3200000 Hz
 * 		    sample loss is to be expected for rates > 2400000
 * \return 0 on success, -EINVAL on invalid rate
 */
extern int rtlsdr_set_sample_rate(rtlsdr_dev_t *dev, uint32_t rate);

/*!
 * Get actual sample rate the device is configured to.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return 0 on error, sample rate in Hz otherwise
 */
extern uint32_t rtlsdr_get_sample_rate(rtlsdr_dev_t *dev);

/*!
 * Enable test mode that returns an 8 bit counter instead of the samples.
 * The counter is generated inside the RTL2832.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param test mode, 1 means enabled, 0 disabled
 * \return 0 on success
 */
extern int rtlsdr_set_testmode(rtlsdr_dev_t *dev, int on);

/*!
 * Enable or disable the internal digital AGC of the RTL2832.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param digital AGC mode, 1 means enabled, 0 disabled
 * \return 0 on success
 */
extern int rtlsdr_set_agc_mode(rtlsdr_dev_t *dev, int on);

/*!
 * Enable or disable the direct sampling mode. When enabled, the IF mode
 * of the RTL2832 is activated, and rtlsdr_set_center_freq() will control
 * the IF-frequency of the DDC, which can be used to tune from 0 to 28.8 MHz
 * (xtal frequency of the RTL2832).
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param on 0 means disabled, 1 I-ADC input enabled, 2 Q-ADC input enabled
 * \return 0 on success
 */
extern int rtlsdr_set_direct_sampling(rtlsdr_dev_t *dev, int on);

/*!
 * Get state of the direct sampling mode
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return -1 on error, 0 means disabled, 1 I-ADC input enabled
 *	    2 Q-ADC input enabled
 */
extern int rtlsdr_get_direct_sampling(rtlsdr_dev_t *dev);

/*!
 * Enable or disable offset tuning for zero-IF tuners, which allows to avoid
 * problems caused by the DC offset of the ADCs and 1/f noise.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param on 0 means disabled, 1 enabled
 * \return 0 on success
 */
extern int rtlsdr_set_offset_tuning(rtlsdr_dev_t *dev, int on);

/*!
 * Get state of the offset tuning mode
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return -1 on error, 0 means disabled, 1 enabled
 */
extern int rtlsdr_get_offset_tuning(rtlsdr_dev_t *dev);

/* streaming functions */

extern int rtlsdr_reset_buffer(rtlsdr_dev_t *dev);

extern int rtlsdr_read_sync(rtlsdr_dev_t *dev, void *buf, int len, int *n_read);

typedef void(*rtlsdr_read_async_cb_t)(unsigned char *buf, uint32_t len, void *ctx);

/*!
 * Read samples from the device asynchronously. This function will block until
 * it is being canceled using rtlsdr_cancel_async()
 *
 * NOTE: This function is deprecated and is subject for removal.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param cb callback function to return received samples
 * \param ctx user specific context to pass via the callback function
 * \return 0 on success
 */
extern int rtlsdr_wait_async(rtlsdr_dev_t *dev, rtlsdr_read_async_cb_t cb, void *ctx);

/*!
 * Read samples from the device asynchronously. This function will block until
 * it is being canceled using rtlsdr_cancel_async()
 *
 * \param dev the device handle given by rtlsdr_open()
 * \param cb callback function to return received samples
 * \param ctx user specific context to pass via the callback function
 * \param buf_num optional buffer count, buf_num * buf_len = overall buffer size
 *		  set to 0 for default buffer count (15)
 * \param buf_len optional buffer length, must be multiple of 512,
 *		  should be a multiple of 16384 (URB size), set to 0
 *		  for default buffer length (16 * 32 * 512)
 * \return 0 on success
 */
extern int rtlsdr_read_async(rtlsdr_dev_t *dev,
				 rtlsdr_read_async_cb_t cb,
				 void *ctx,
				 uint32_t buf_num,
				 uint32_t buf_len);

/*!
 * Cancel all pending asynchronous operations on the device.
 *
 * \param dev the device handle given by rtlsdr_open()
 * \return 0 on success
 */
extern int rtlsdr_cancel_async(rtlsdr_dev_t *dev);

#ifdef __cplusplus
}
#endif

#endif /* __RTL_SDR_H */
