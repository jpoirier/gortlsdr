# Description

gortlsdr is a simple Go interface to devices supported by the RTL-SDR project, which turns certain USB DVB-T dongles
employing the Realtek RTL2832U chipset into a low-cost, general purpose software-defined radio receivers. It wraps all
the functions in the [librtlsdr library](http://sdr.osmocom.org/trac/wiki/rtl-sdr) (including asynchronous read support).

Supported Platforms:

- Linux
- OS X
- Windows


# Installation

## Dependencies
- [Go tools](http://golang.org)
- [librtlsdr] (http://sdr.osmocom.org/trac/wiki/rtl-sdr) - builds dated after 5/5/12

## Usage
All functions in librtlsdr are accessible in the gortlsdr package, but note that ReadAsync can cause a segfault due to
a known issue with libusb_handle_events_timeout in libusb.

    go get -v github.com/jpoirier/gortlsdr

## Examples
See the rtlsdr_eample.go file.

	go run rtlsdr_eample.go


# Credit
The great readme description was copied from the Python wrapper page [pyrtlsdr](https://github.com/roger-/pyrtlsdr/tree/master/rtlsdr).

-joe



