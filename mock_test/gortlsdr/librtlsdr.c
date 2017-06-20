/*
 * Copyright (c) 2015-2017 Joseph D Poirier
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
#include <pthread.h>
#include <unistd.h>

#include "rtl-sdr_moc.h"

#define DEBVICE_CNT (3)
#define DEVICE_GAIN_CNT (29)
#define EEPROM_SIZE (256)
#define DEFAULT_BUF_NUMBER (15)
#define DEFAULT_BUF_LENGTH (16 * 32 * 512)
#define STRINGS_OFFSET_START (9)
#define MAX_RAW_STR_SZ (2*35+2)


static void do_init(void);
static bool dev_valid(rtlsdr_dev_t *dev);

enum rtlsdr_async_status {
	RTLSDR_INACTIVE = 0,
	RTLSDR_CANCELING,
	RTLSDR_RUNNING
};

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
	uint32_t xfer_buf_num;
	uint32_t xfer_buf_len;
	enum rtlsdr_tuner type;
	enum rtlsdr_async_status async_status;
	rtlsdr_read_async_cb_t cb;
	int async_cancel;
	void *cb_ctx;
	pthread_mutex_t lock;
	char eeprom_buffer[256];
	int gains[DEVICE_GAIN_CNT];
	char xbuf[DEFAULT_BUF_LENGTH];
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

static struct rtlsdr_dev s0 = {0};
static struct rtlsdr_dev s1 = {0};
static struct rtlsdr_dev s2 = {0};

struct rtlsdr_dev *rtlsdr_devs[DEBVICE_CNT] = {&s0, &s1, &s2};

static bool is_initialized = false;

static pthread_t tid;

void do_init(void) {
	if (is_initialized)
		return;

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
		.eeprom_buffer = {0x28, 0x32, 0x09, 0x01, 0x01, 0x01, 0xA5, 0x03, 0x00,
		0x10, 0x03, 'R', 0x00, 'E', 0x00, 'A', 0x00, 'L', 0x00, 'T', 0x00, 'E', 0x00, 'K', 0x00,
		0x10, 0x03, 'N', 0x00, 'O', 0x00, 'O', 0x00, 'E', 0x00, 'L', 0x00, 'E', 0x00, 'C', 0x00,
		0x0A, 0x03, '1', 0x00, '9', 0x00, '9', 0x00, '1', 0x00},
		.gains = { 0, 9, 14, 27, 37, 77, 87, 125, 144, 157,
				     166, 197, 207, 229, 254, 280, 297, 328,
				     338, 364, 372, 386, 402, 421, 434, 439,
				     445, 480, 496}
	};

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
		.eeprom_buffer = {0x28, 0x32, 0x09, 0x01, 0x01, 0x01, 0xA5, 0x03, 0x00,
		0x10, 0x03, 'R', 0x00, 'E', 0x00, 'A', 0x00, 'L', 0x00, 'T', 0x00, 'E', 0x00, 'K', 0x00,
		0x10, 0x03, 'N', 0x00, 'O', 0x00, 'O', 0x00, 'E', 0x00, 'L', 0x00, 'E', 0x00, 'C', 0x00,
		0x0A, 0x03, '2', 0x00, '9', 0x00, '9', 0x00, '2', 0x00},
		.gains = { 0, 9, 14, 27, 37, 77, 87, 125, 144, 157,
					166, 197, 207, 229, 254, 280, 297, 328,
					338, 364, 372, 386, 402, 421, 434, 439,
					445, 480, 496}
	};

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
		.eeprom_buffer = {0x28, 0x32, 0x09, 0x01, 0x01, 0x01, 0xA5, 0x03, 0x00,
		0x10, 0x03, 'R', 0x00, 'E', 0x00, 'A', 0x00, 'L', 0x00, 'T', 0x00, 'E', 0x00, 'K', 0x00,
		0x10, 0x03, 'N', 0x00, 'O', 0x00, 'O', 0x00, 'E', 0x00, 'L', 0x00, 'E', 0x00, 'C', 0x00,
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
	if (!dev || !dev_valid(dev))
		return -1;

	dev->ppm = ppm;
	return 0;
}

int rtlsdr_set_xtal_freq(rtlsdr_dev_t *dev, uint32_t rtl_freq, uint32_t tuner_freq) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->tuner_freq = tuner_freq;
	dev->rtl_freq = rtl_freq;
	return 0;
}

int rtlsdr_get_xtal_freq(rtlsdr_dev_t *dev, uint32_t *rtl_freq, uint32_t *tuner_freq) {
	if (!dev || !dev_valid(dev))
		return -1;

	*rtl_freq = dev->rtl_freq;
	*tuner_freq = dev->tuner_freq;
	return 0;
}

int rtlsdr_get_usb_strings(rtlsdr_dev_t *dev, char *manufact, char *product, char *serial) {
	if (!dev || !dev_valid(dev))
		return -1;

	char *p;
	if (dev == &s0) {
		p = &(s0.eeprom_buffer[STRINGS_OFFSET_START]);
	} else if (dev == &s1) {
		p = &(s1.eeprom_buffer[STRINGS_OFFSET_START]);
	} else if (dev == &s2) {
		p = &(s2.eeprom_buffer[STRINGS_OFFSET_START]);
	} else {
		return -2;
	}

	int i, j, sz;

	sz = p[0];
	if (manufact) {
		for (i = 0, j = 0; j < (sz-2); i++, j+=2) {
			manufact[i] = p[2+j];
		}
		manufact[i] = '\0';
	}

	p += sz;
	sz = p[0];
	if (product) {
		for (i = 0, j = 0; j < (sz-2); i++, j+=2) {
			product[i] = p[2+j];
		}
		product[i] = '\0';
	}

	p += sz;
	sz = p[0];
	if (serial) {
		for (i = 0, j = 0; j < (sz-2); i++, j+=2) {
			serial[i] = p[2+j];
		}
		serial[i] = '\0';

	}

	return 0;
}

int rtlsdr_write_eeprom(rtlsdr_dev_t *dev, uint8_t *data, uint8_t offset, uint16_t len) {
	if (!dev || !dev_valid(dev))
		return -1;

	if (len > (EEPROM_SIZE-offset))
		return -2;

	if (memcpy(&(dev->eeprom_buffer[offset]), data, len) == 0)
		return -3;

	return 0;
}

int rtlsdr_read_eeprom(rtlsdr_dev_t *dev, uint8_t *data, uint8_t offset, uint16_t len) {
	if (!dev || !dev_valid(dev))
		return -1;

	if (len > (EEPROM_SIZE-offset))
		return -2;

	char *p;
	if (dev == &s0) {
		p = &(s0.eeprom_buffer[offset]);
	} else if (dev == &s1) {
		p = &(s1.eeprom_buffer[offset]);
	} else if (dev == &s2) {
		p = &(s2.eeprom_buffer[offset]);
	} else {
		return -3;
	}

	if (memcpy(data, p, len) == 0)
		return -1;

	return 0;
}

int rtlsdr_set_center_freq(rtlsdr_dev_t *dev, uint32_t freq) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->center_freq = freq;
	return 0;
}

uint32_t rtlsdr_get_center_freq(rtlsdr_dev_t *dev) {
	if (!dev || !dev_valid(dev))
		return -1;

	return dev->center_freq;
}

int rtlsdr_get_freq_correction(rtlsdr_dev_t *dev) {
	if (!dev || !dev_valid(dev))
		return -1;

	return dev->ppm;
}

enum rtlsdr_tuner rtlsdr_get_tuner_type(rtlsdr_dev_t *dev){
	if (!dev || !dev_valid(dev))
		return RTLSDR_TUNER_UNKNOWN;

	return dev->type;
}

int rtlsdr_get_tuner_gains(rtlsdr_dev_t *dev, int *gains) {
	if (!dev || !dev_valid(dev))
		return -1;

	gains = &(dev->gains[0]);
	return (DEVICE_GAIN_CNT / sizeof(int));
}

int rtlsdr_set_tuner_bandwidth(rtlsdr_dev_t *dev, uint32_t bw) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->tuner_bandwidth = bw;
	return 0;
}

int rtlsdr_set_tuner_gain(rtlsdr_dev_t *dev, int gain) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->gain = gain;
	return 0;
}

int rtlsdr_get_tuner_gain(rtlsdr_dev_t *dev) {
	if (!dev || !dev_valid(dev))
		return -1;

	return dev->gain;
}

int rtlsdr_set_tuner_if_gain(rtlsdr_dev_t *dev, int stage, int gain) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->gain = gain;
	return 0;
}

int rtlsdr_set_tuner_gain_mode(rtlsdr_dev_t *dev, int mode) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->gain_mode = mode;
	return 0;
}

int rtlsdr_set_sample_rate(rtlsdr_dev_t *dev, uint32_t samp_rate) {
	if (!dev || !dev_valid(dev))
		return -1;

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
	if (!dev || !dev_valid(dev))
		return -1;

	return dev->sample_rate;
}

int rtlsdr_set_testmode(rtlsdr_dev_t *dev, int on) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->test_mode = on;
	return 0;
}

int rtlsdr_set_agc_mode(rtlsdr_dev_t *dev, int on) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->agc_mode = on;
	return 0;
}

int rtlsdr_set_direct_sampling(rtlsdr_dev_t *dev, int on) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->direct_sampling_mode = on;
	return 0;
}

int rtlsdr_get_direct_sampling(rtlsdr_dev_t *dev){
	if (!dev || !dev_valid(dev))
		return -1;

	return dev->direct_sampling_mode;
}

int rtlsdr_set_offset_tuning(rtlsdr_dev_t *dev, int on) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->offset_tuning = on;
	return 0;
}

int rtlsdr_get_offset_tuning(rtlsdr_dev_t *dev) {
	if (!dev || !dev_valid(dev))
		return -1;

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

	rtlsdr_dev_t *dev;
	if (index == 0) {
		dev = &s0;
	} else if (index == 1) {
		dev = &s1;
	} else if (index == 2) {
		dev = &s2;
	} else {
		return -1;
	}

	return rtlsdr_get_usb_strings(dev, manufact, product, serial);
}

int rtlsdr_get_index_by_serial(const char *serial) {
	do_init();

	if (!serial)
		return -1;

	char s[40];

	int e = rtlsdr_get_usb_strings(&s0, 0, 0, &s[0]);
	if (e != 0)
		return e;
	if (!strcmp(serial, &s[0])) {
		return 0;
	}

	e = rtlsdr_get_usb_strings(&s1, 0, 0, &s[0]);
	if (e != 0)
		return e;
	if (!strcmp(serial, &s[0])) {
		return 0;
	}

	e = rtlsdr_get_usb_strings(&s2, 0, 0, &s[0]);
	if (e != 0)
		return e;
	if (!strcmp(serial, &s[0])) {
		return 0;
	}

	return -2;
}

int rtlsdr_open(rtlsdr_dev_t **out_dev, uint32_t index) {
	do_init();

	if (!out_dev)
		return -1;

	// FIXME:
	// if (pthread_mutex_init(&lock, NULL) != 0) {
	// 	return 1;
	// }

	int status = 1;
	switch(index) {
	case 0:
		*out_dev = &s0;
		s0.status = true;
		s0.async_status = RTLSDR_INACTIVE;
		pthread_mutex_init(&s0.lock, NULL);
		status = 0;
		break;
	case 1:
		*out_dev = &s1;
		s1.status = true;
		s1.async_status = RTLSDR_INACTIVE;
		pthread_mutex_init(&s1.lock, NULL);
		status = 0;
		break;
	case 2:
		*out_dev = &s2;
		s2.status = true;
		s2.async_status = RTLSDR_INACTIVE;
		pthread_mutex_init(&s2.lock, NULL);
		status = 0;
		break;
	}

	return status;
}

int rtlsdr_close(rtlsdr_dev_t *dev) {
	if (!dev || !dev_valid(dev))
		return -1;

	dev->status = false;
	return 0;
}

int rtlsdr_reset_buffer(rtlsdr_dev_t *dev) {
	if (!dev || !dev_valid(dev))
		return -1;

	return 0;
}

int rtlsdr_read_sync(rtlsdr_dev_t *dev, void *buf, int len, int *n_read) {
	if (!dev || !dev_valid(dev))
		return -1;

	if (len > DEFAULT_BUF_LENGTH)
		len = DEFAULT_BUF_LENGTH;
	*n_read = len;

	if (dev == &s0) {
		memcpy(buf, &s0.xbuf[0], len);
	} else if (dev == &s1) {
		memcpy(buf, &s1.xbuf[0], len);
	} else if (dev == &s2) {
		memcpy(buf, &s2.xbuf[0], len);
	} else {
		return -2;
	}

	return 0;
}

static unsigned char async_buf[DEFAULT_BUF_LENGTH];
void read_asyn_handler(void *dev) {
	struct rtlsdr_dev *d = (struct rtlsdr_dev*)dev;
	rtlsdr_read_async_cb_t cb = d->cb;
	// void *cb_ctx = d->cb_ctx;

	for (;;) {
		pthread_mutex_lock(&d->lock);
		if (d->async_status == RTLSDR_CANCELING) {
			pthread_mutex_unlock(&d->lock);
			break;
		}
		pthread_mutex_unlock(&d->lock);
		cb(async_buf, DEFAULT_BUF_LENGTH, NULL);
		sleep(1);
	}
	d->async_status = RTLSDR_INACTIVE;
}

int rtlsdr_read_async(rtlsdr_dev_t *dev, rtlsdr_read_async_cb_t cb, void *ctx,
			  uint32_t buf_num, uint32_t buf_len) {
	// unsigned int i;
	// int r = 0;
	// struct timeval tv = { 1, 0 };
	// struct timeval zerotv = { 0, 0 };
	// enum rtlsdr_async_status next_status = RTLSDR_INACTIVE;

	if (!dev || !dev_valid(dev))
		return -1;

	if (RTLSDR_INACTIVE != dev->async_status)
		return -2;

	dev->async_status = RTLSDR_RUNNING;
	dev->async_cancel = 0;

	dev->cb = cb;
	dev->cb_ctx = ctx;

	if (buf_num > 0)
		dev->xfer_buf_num = buf_num;
	else
		dev->xfer_buf_num = DEFAULT_BUF_NUMBER;

	// if (buf_len > 0 && buf_len % 512 == 0)  len must be multiple of 512
	// 	dev->xfer_buf_len = buf_len;
	// else
		dev->xfer_buf_len = DEFAULT_BUF_LENGTH;

	if (pthread_create(&tid, NULL, (void*)read_asyn_handler, (void*)dev) != 0)
		return -3;

	return 0;
}

int rtlsdr_cancel_async(rtlsdr_dev_t *dev) {
	if (!dev || !dev_valid(dev))
		return -1;

	// if streaming, try to cancel gracefully
	pthread_mutex_lock(&dev->lock);
	if (RTLSDR_RUNNING == dev->async_status) {
		dev->async_status = RTLSDR_CANCELING;
		dev->async_cancel = 1;
	}
	pthread_mutex_unlock(&dev->lock);
	return 0;
}

int rtlsdr_set_bias_tee(rtlsdr_dev_t *dev, int on) {
	if (!dev)
		return -1;
	return 0;
}
