# Description

gortlsdr is a simple Go interface to devices supported by the RTL-SDR project, which turns certain USB DVB-T dongles
employing the Realtek RTL2832U chipset into low-cost, general purpose software-defined radio receivers. It wraps all the
functions in the [librtlsdr library](http://sdr.osmocom.org/trac/wiki/rtl-sdr) (including asynchronous read support).

# Usage

All functions in librtlsdr are accessible in the gortlsdr package except rtlsdr_read_async, which can cause segfaults
due to a known issue calling libusb_handle_events_timeout in libusb.
Some documentation can be found in docstrings in the latter file.

    go get -v github.com/kylelemons/gousb/lsusb

## Examples

See the rtlsdr_eample.go file.


# Dependencies

* Windows/Linux/OSX
* [Go tools](http://golang.org)
* [librtlsdr] (http://sdr.osmocom.org/trac/wiki/rtl-sdr) - builds dated after 5/5/12


# Credit
The great readme description was copied from the Python wrapper page [pyrtlsdr](https://github.com/roger-/pyrtlsdr/tree/master/rtlsdr).

-joe



