# Description

***** The code in this project is under development and is not complete nor has it been fully tested! *****

gortlsdr is a simple Go interface to devices supported by the RTL-SDR project, which turns certain USB DVB-T dongles
employing the Realtek RTL2832U chipset into low-cost, general purpose software-defined radio receivers. It wraps all the
functions in the [librtlsdr library](http://sdr.osmocom.org/trac/wiki/rtl-sdr) (including asynchronous read support).

# Usage

All functions in librtlsdr are accessible via the gortlsdr package.
Some documentation can be found in docstrings in the latter file.

## Examples

Simple way to read and print some samples:


# configure device


# configure device



# Dependencies

* Windows/Linux/OSX
* [Go tools](http://golang.org)
* librtlsdr (builds dated after 5/5/12)


# Troubleshooting

* Some operating systems (Linux, OS X) seem to result in libusb buffer issues when performing small reads. Try reading 1024
(or higher powers of two) samples at a time if you have problems.
* If you're having librtlsdr import errors in Windows, make sure all the DLL files are in your system path, or the same folder
as this README file. Also make sure you have all of *their* dependencies (e.g. the Visual Studio runtime files). If rtl_sdr.exe
works, then you should be okay.
* In Windows, you can't mix the 64 bit version of Go with 32 bit builds of librtlsdr.

# Credit
Credit to Roger for his Python wrapper [pyrtlsdr](https://github.com/roger-/pyrtlsdr/tree/master/rtlsdr) and the great readme file!

-joe
