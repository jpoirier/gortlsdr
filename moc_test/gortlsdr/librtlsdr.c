/*
 * Copyright (c) 2015-2016 Joseph D Poirier
 * Distributable under the terms of The New BSD License
 * that can be found in the LICENSE file.
 *
 * $ gcc -Wall -c librtlsdr_moc.c -o librtlsdr_moc.o
 * $ ar rcs librtlsdr_moc.a librtlsdr_moc.o
 *
 */
#include <errno.h>
#include <signal.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <stdint.h>

#include "rtl-sdr_moc.h"

#define DEBVICE_CNT (3)
#define DEVICE_GAIN_CNT (29)

// struct rtlsdr_dev {
// 	libusb_context *ctx;
// 	struct libusb_device_handle *devh;
// 	uint32_t xfer_buf_num;
// 	uint32_t xfer_buf_len;
// 	struct libusb_transfer **xfer;
// 	unsigned char **xfer_buf;
// 	rtlsdr_read_async_cb_t cb;
// 	void *cb_ctx;
// 	enum rtlsdr_async_status async_status;
// 	int async_cancel;
// 	/* rtl demod context */
// 	uint32_t rate; /* Hz */
// 	uint32_t rtl_xtal; /* Hz */
// 	int fir[FIR_LEN];
// 	int direct_sampling;
// 	/* tuner context */
// 	enum rtlsdr_tuner tuner_type;
// 	rtlsdr_tuner_iface_t *tuner;
// 	uint32_t tun_xtal; /* Hz */
// 	uint32_t freq; /* Hz */
// 	uint32_t bw;
// 	uint32_t offs_freq; /* Hz */
// 	int corr; /* ppm */
// 	int gain; /* tenth dB */
// 	struct e4k_state e4k_s;
// 	struct r82xx_config r82xx_c;
// 	struct r82xx_priv r82xx_p;
// 	/* status */
// 	int dev_lost;
// 	int driver_active;
// 	unsigned int xfer_errors;
// };

struct rtlsdr_dev {
	bool status;
	int ppm;
	uint32_t tuner_freq;
	uint32_t rtl_freq;
	uint32_t center_freq;
	uint32_t tuner_bandwidth;
	int if_gain;
	int gain;
	int gain_mode;
	uint32_t sample_rate;
	int test_mode;
	int agc_mode;
	int direct_sampling_mode;
	int offset_tuning;
	enum rtlsdr_tuner type;
	char eeprom_buffer[256];
	char manufact[37]; 	// fat pointer
	char product[37]; 	// fat pointer
	char serial[37];	// fat pointer
	int gains[DEVICE_GAIN_CNT];
};

/*
   0  |  1   | 2(L) 3(U) | 4(L)  5(U) |   6 (0xA5)  |   7 (|0x01)   | 7 (|0x02) |
 0x28 | 0x32 | Vendor ID | Product ID | Have Serial | Remote Wakeup | Enable IR |

string start offset index is 9 and indicates the string size in bytes
max string length is 35 * 2 + 2 header bytes = 72 bytes
header byte [0] is the total string length <= 72 (info string length <= 72 - 2 header bytes)
header byte [1] is 0x03
Manufact info string start is [3], info string end is [header byte [0] - 2]
Product info string start [], info string end is [header byte [] - 2]
Serial info string start [], info string end is [header byte [] - 2]

strings are written as "["char", 0x00]" pairs
*/

static int device_count = DEBVICE_CNT;

struct rtlsdr_dev s0 = {0};
struct rtlsdr_dev s1 = {0};
struct rtlsdr_dev s2 = {0};

struct rtlsdr_dev *rtlsdr_devs[DEBVICE_CNT] = {&s0, &s1, &s2};

static bool is_initialized = false;

void do_init(void) {
	if (is_initialized)
		return;

	s0.manufact[0] = 7;
	strcpy(&(s0.manufact[1]), "REALTEK\0");
	s0.product[0] = 7;
	strcpy(&(s0.product[1]), "NOOELEC\0") ;
	s0.serial[0] = 4;
	strcpy(&(s0.serial[1]), "1991\0");
	s0 = (struct rtlsdr_dev){
		.status = false,
		.ppm = 50,
		.tuner_freq = 1700000,
		.rtl_freq = 1700000,
		.center_freq = 1700000,
		.tuner_bandwidth = 1000000,
		.if_gain = 50,
		.gain = 100,
		.gain_mode = 1,
		.sample_rate = 3200000,
		.test_mode = 0,
		.agc_mode = 0,
		.direct_sampling_mode = 1,
		.offset_tuning = 1,
		.type = RTLSDR_TUNER_R828D,
		.eeprom_buffer = {0x28, 0x32, 0x09, 0x01, 0x01, 0x01, 0xA5, 0x03,
		0x1F, 0x03, 'R', 0x00, 'E', 0x00, 'A', 0x00, 'L', 0x00, 'T', 0x00, 'E', 0x00, 'K', 0x00,
		0x1F, 0x03, 'N', 0x00, 'O', 0x00, 'O', 0x00, 'E', 0x00, 'L', 0x00, 'E', 0x00, 'C', 0x00,
		0x0A, 0x03, '1', 0x00, '9', 0x00, '9', 0x00, '1', 0x00},
		.gains = { 0, 9, 14, 27, 37, 77, 87, 125, 144, 157,
				     166, 197, 207, 229, 254, 280, 297, 328,
				     338, 364, 372, 386, 402, 421, 434, 439,
				     445, 480, 496}
	};

	s1.manufact[0] = 7;
	strcpy(&(s0.manufact[1]), "REALTEK\0");
	s1.product[0] = 7;
	strcpy(&(s0.product[1]), "NOOELEC\0") ;
	s1.serial[0] = 4;
	strcpy(&(s0.serial[1]), "2992\0");
	s1 = (struct rtlsdr_dev){
		.status = false,
		.ppm = 51,
		.tuner_freq = 1700001,
		.rtl_freq = 1700001,
		.center_freq = 1700001,
		.tuner_bandwidth = 1000001,
		.if_gain = 51,
		.gain = 101,
		.gain_mode = 0,
		.sample_rate = 3200000,
		.test_mode = 1,
		.agc_mode = 1,
		.direct_sampling_mode = 0,
		.offset_tuning = 0,
		.type = RTLSDR_TUNER_R820T,
		.eeprom_buffer = {0x28, 0x32, 0x09, 0x01, 0x01, 0x01, 0xA5, 0x03,
		0x1F, 0x03, 'R', 0x00, 'E', 0x00, 'A', 0x00, 'L', 0x00, 'T', 0x00, 'E', 0x00, 'K', 0x00,
		0x1F, 0x03, 'N', 0x00, 'O', 0x00, 'O', 0x00, 'E', 0x00, 'L', 0x00, 'E', 0x00, 'C', 0x00,
		0x0A, 0x03, '2', 0x00, '9', 0x00, '9', 0x00, '2', 0x00},
		.gains = { 0, 9, 14, 27, 37, 77, 87, 125, 144, 157,
				     166, 197, 207, 229, 254, 280, 297, 328,
				     338, 364, 372, 386, 402, 421, 434, 439,
				     445, 480, 496}
	};

	s2.manufact[0] = 7;
	strcpy(&(s0.manufact[1]), "REALTEK\0");
	s2.product[0] = 7;
	strcpy(&(s0.product[1]), "NOOELEC\0") ;
	s2.serial[0] = 4;
	strcpy(&(s0.serial[1]), "3993\0");
	s2 = (struct rtlsdr_dev){
		.status = false,
		.ppm = 52,
		.tuner_freq = 1700002,
		.rtl_freq = 1700002,
		.center_freq = 1700002,
		.tuner_bandwidth = 1000002,
		.if_gain = 52,
		.gain = 102,
		.gain_mode = 1,
		.sample_rate = 3200000,
		.test_mode = 0,
		.agc_mode = 1,
		.direct_sampling_mode = 1,
		.offset_tuning = 1,
		.type = RTLSDR_TUNER_E4000,
		.eeprom_buffer = {0x28, 0x32, 0x09, 0x01, 0x01, 0x01, 0xA5, 0x03,
		0x1F, 0x03, 'R', 0x00, 'E', 0x00, 'A', 0x00, 'L', 0x00, 'T', 0x00, 'E', 0x00, 'K', 0x00,
		0x1F, 0x03, 'N', 0x00, 'O', 0x00, 'O', 0x00, 'E', 0x00, 'L', 0x00, 'E', 0x00, 'C', 0x00,
		0x0A, 0x03, '3', 0x00, '9', 0x00, '9', 0x00, '3', 0x00},
		.gains = { 0, 9, 14, 27, 37, 77, 87, 125, 144, 157,
				     166, 197, 207, 229, 254, 280, 297, 328,
				     338, 364, 372, 386, 402, 421, 434, 439,
				     445, 480, 496 }
	};
	is_initialized = true;
}

bool dev_valid(rtlsdr_dev_t *dev) {
	if (dev == &s0 || dev == &s1 || dev == &s2) {
		return true;
	}
	return false;
}

int rtlsdr_set_freq_correction(rtlsdr_dev_t *dev, int ppm) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->ppm = ppm;
	return 0;
}

int rtlsdr_set_xtal_freq(rtlsdr_dev_t *dev, uint32_t rtl_freq, uint32_t tuner_freq) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->tuner_freq = tuner_freq;
	dev->rtl_freq = rtl_freq;
	return 0;
}

int rtlsdr_get_xtal_freq(rtlsdr_dev_t *dev, uint32_t *rtl_freq, uint32_t *tuner_freq) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	*rtl_freq = dev->rtl_freq;
	*tuner_freq = dev->tuner_freq;
	return 0;
}

int rtlsdr_get_usb_strings(rtlsdr_dev_t *dev, char *manufact, char *product, char *serial) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	if (dev == &s0) {
		memcpy(manufact, &s0.manufact[1], s0.manufact[0]);
		memcpy(product, &s0.product[1], s0.product[0]);
		memcpy(serial, &s0.serial[1], s0.serial[0]);
	} else if (dev == &s1) {
		memcpy(manufact, &s1.manufact[1], s1.manufact[0]);
		memcpy(product, &s1.product[1], s1.product[0]);
		memcpy(serial, &s1.serial[1], s1.serial[0]);
	} else if (dev == &s1) {
		memcpy(manufact, &s2.manufact[1], s2.manufact[0]);
		memcpy(product, &s2.product[1], s2.product[0]);
		memcpy(serial, &s2.serial[1], s2.serial[0]);
	} else {
		return -3;
	}

	return 0;
}

// TODO:
int rtlsdr_write_eeprom(rtlsdr_dev_t *dev, uint8_t *data, uint8_t offset, uint16_t len) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	// if ((len + offset) > 256)
	// 	return -2;
	// int i;
	// for (i = 0; i < len; i++)
	// 	eeprom_buffer[i+offset] = data[i];
	return 0;
}

// TODO:
int rtlsdr_read_eeprom(rtlsdr_dev_t *dev, uint8_t *data, uint8_t offset, uint16_t len) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	// if ((len + offset) > 256)
	// 	return -2;
	// int i;
	// for (i = 0; i < len; i++)
	// 	data[i] = eeprom_buffer[i+offset];

	return 0;
}

int rtlsdr_set_center_freq(rtlsdr_dev_t *dev, uint32_t freq) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->center_freq = freq;
	return 0;
}

uint32_t rtlsdr_get_center_freq(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return dev->center_freq;
}

int rtlsdr_get_freq_correction(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return dev->ppm;
}

enum rtlsdr_tuner rtlsdr_get_tuner_type(rtlsdr_dev_t *dev){
	if (!dev)
		return RTLSDR_TUNER_UNKNOWN;
	if (!dev_valid(dev))
		return -2;

	return dev->type;
}

int rtlsdr_get_tuner_gains(rtlsdr_dev_t *dev, int *gains) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	gains = &(dev->gains[0]);
	return (DEVICE_GAIN_CNT / sizeof(int));
}

int rtlsdr_set_tuner_bandwidth(rtlsdr_dev_t *dev, uint32_t bw) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->tuner_bandwidth = bw;
	return 0;
}

int rtlsdr_set_tuner_gain(rtlsdr_dev_t *dev, int gain) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->gain = gain;
	return 0;
}

int rtlsdr_get_tuner_gain(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return dev->gain;
}

int rtlsdr_set_tuner_if_gain(rtlsdr_dev_t *dev, int stage, int gain) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->gain = gain;
	return 0;
}

int rtlsdr_set_tuner_gain_mode(rtlsdr_dev_t *dev, int mode) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->gain_mode = mode;
	return 0;
}

int rtlsdr_set_sample_rate(rtlsdr_dev_t *dev, uint32_t samp_rate) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	/* check if the rate is supported by the resampler */
	if ((samp_rate <= 225000) || (samp_rate > 3200000) ||
	   ((samp_rate > 300000) && (samp_rate <= 900000))) {
		fprintf(stderr, "Invalid sample rate: %u Hz\n", samp_rate);
		return -EINVAL;
	}
	dev->sample_rate = samp_rate;
	return 0;
}

uint32_t rtlsdr_get_sample_rate(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return dev->sample_rate;
}

int rtlsdr_set_testmode(rtlsdr_dev_t *dev, int on) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->test_mode = on;
	return 0;
}

int rtlsdr_set_agc_mode(rtlsdr_dev_t *dev, int on) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->agc_mode = on;
	return 0;
}

int rtlsdr_set_direct_sampling(rtlsdr_dev_t *dev, int on) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->direct_sampling_mode = on;
	return 0;
}

int rtlsdr_get_direct_sampling(rtlsdr_dev_t *dev){
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return dev->direct_sampling_mode;
}

int rtlsdr_set_offset_tuning(rtlsdr_dev_t *dev, int on) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->offset_tuning = on;
	return 0;
}

int rtlsdr_get_offset_tuning(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return dev->offset_tuning;
}

uint32_t rtlsdr_get_device_count(void) {
	return device_count;
}

const char *rtlsdr_get_device_name(uint32_t index) {
	do_init();
	if (index < 0 || index > 2)
		return "";

	return "Generic RTL2832U OEM";
}

int rtlsdr_get_device_usb_strings(uint32_t index, char *manufact, char *product, char *serial) {
	do_init();

	if (index == 0) {
		memcpy(manufact, &s0.manufact[1], s0.manufact[0]);
		memcpy(product, &s0.product[1], s0.product[0]);
		memcpy(serial, &s0.serial[1], s0.serial[0]);
	} else if (index == 1) {
		memcpy(manufact, &s1.manufact[1], s1.manufact[0]);
		memcpy(product, &s1.product[1], s1.product[0]);
		memcpy(serial, &s1.serial[1], s1.serial[0]);
	} else if (index == 2) {
		memcpy(manufact, &s2.manufact[1], s2.manufact[0]);
		memcpy(product, &s2.product[1], s2.product[0]);
		memcpy(serial, &s2.serial[1], s2.serial[0]);
	} else {
		return -1;
	}

	return 0;
}

int rtlsdr_get_index_by_serial(const char *serial) {
	do_init();

	if (!serial)
		return -1;

	if (!strcmp(serial, &(s0.serial[1]))) {
		return 0;
	} else if (!strcmp(serial, &(s1.serial[1]))) {
		return 1;
	} else if (!strcmp(serial, &(s1.serial[1]))) {
		return 2;
	}

	return -2;
}

int rtlsdr_open(rtlsdr_dev_t **out_dev, uint32_t index) {
	do_init();

	if (!out_dev)
		return -1;

	int status = 1;
	switch(index) {
	case 0:
		*out_dev = &s0;
		s0.status = true;
		status = 0;
		break;
	case 1:
		*out_dev = &s1;
		s1.status = true;
		status = 0;
		break;
	case 2:
		*out_dev = &s1;
		s2.status = true;
		status = 0;
		break;
	}

	return status;
}

int rtlsdr_close(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	dev->status = false;
	return 0;
}

// TODO:
int rtlsdr_reset_buffer(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return 0;
}

// TODO:
int rtlsdr_read_sync(rtlsdr_dev_t *dev, void *buf, int len, int *n_read) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	return 0;
}

// TODO:
int rtlsdr_read_async(rtlsdr_dev_t *dev, rtlsdr_read_async_cb_t cb, void *ctx,
			  uint32_t buf_num, uint32_t buf_len) {
	// unsigned int i;
	// int r = 0;
	// struct timeval tv = { 1, 0 };
	// struct timeval zerotv = { 0, 0 };
	// enum rtlsdr_async_status next_status = RTLSDR_INACTIVE;

	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	// if (RTLSDR_INACTIVE != dev->async_status)
	// 	return -2;

	// dev->async_status = RTLSDR_RUNNING;
	// dev->async_cancel = 0;

	// dev->cb = cb;
	// dev->cb_ctx = ctx;

	// if (buf_num > 0)
	// 	dev->xfer_buf_num = buf_num;
	// else
	// 	dev->xfer_buf_num = DEFAULT_BUF_NUMBER;

	// if (buf_len > 0 && buf_len % 512 == 0) /* len must be multiple of 512 */
	// 	dev->xfer_buf_len = buf_len;
	// else
	// 	dev->xfer_buf_len = DEFAULT_BUF_LENGTH;

	// _rtlsdr_alloc_async_buffers(dev);

	// for(i = 0; i < dev->xfer_buf_num; ++i) {
	// 	libusb_fill_bulk_transfer(dev->xfer[i],
	// 				  dev->devh,
	// 				  0x81,
	// 				  dev->xfer_buf[i],
	// 				  dev->xfer_buf_len,
	// 				  _libusb_callback,
	// 				  (void *)dev,
	// 				  BULK_TIMEOUT);

	// 	r = libusb_submit_transfer(dev->xfer[i]);
	// 	if (r < 0) {
	// 		fprintf(stderr, "Failed to submit transfer %i!\n", i);
	// 		dev->async_status = RTLSDR_CANCELING;
	// 		break;
	// 	}
	// }

	// while (RTLSDR_INACTIVE != dev->async_status) {
	// 	r = libusb_handle_events_timeout_completed(dev->ctx, &tv,
	// 						   &dev->async_cancel);
	// 	if (r < 0) {
	// 		/*fprintf(stderr, "handle_events returned: %d\n", r);*/
	// 		if (r == LIBUSB_ERROR_INTERRUPTED) /* stray signal */
	// 			continue;
	// 		break;
	// 	}

	// 	if (RTLSDR_CANCELING == dev->async_status) {
	// 		next_status = RTLSDR_INACTIVE;

	// 		if (!dev->xfer)
	// 			break;

	// 		for(i = 0; i < dev->xfer_buf_num; ++i) {
	// 			if (!dev->xfer[i])
	// 				continue;

	// 			if (LIBUSB_TRANSFER_CANCELLED !=
	// 					dev->xfer[i]->status) {
	// 				r = libusb_cancel_transfer(dev->xfer[i]);
	// 				/* handle events after canceling
	// 				 * to allow transfer status to
	// 				 * propagate */
	// 				libusb_handle_events_timeout_completed(dev->ctx,
	// 								       &zerotv, NULL);
	// 				if (r < 0)
	// 					continue;

	// 				next_status = RTLSDR_CANCELING;
	// 			}
	// 		}

	// 		if (dev->dev_lost || RTLSDR_INACTIVE == next_status) {
	// 			/* handle any events that still need to
	// 			 * be handled before exiting after we
	// 			 * just cancelled all transfers */
	// 			libusb_handle_events_timeout_completed(dev->ctx,
	// 							       &zerotv, NULL);
	// 			break;
	// 		}
	// 	}
	// }

	// _rtlsdr_free_async_buffers(dev);

	// dev->async_status = next_status;

	return 0;
}

// TODO:
int rtlsdr_cancel_async(rtlsdr_dev_t *dev) {
	if (!dev)
		return -1;
	if (!dev_valid(dev))
		return -2;

	// /* if streaming, try to cancel gracefully */
	// if (RTLSDR_RUNNING == dev->async_status) {
	// 	dev->async_status = RTLSDR_CANCELING;
	// 	dev->async_cancel = 1;
	// 	return 0;
	// }
	return 0;
}
