[![gortlsdr version](https://img.shields.io/github/tag/jpoirier/gortlsdr.svg?style=flat&label=gortlsdr)](https://github.com/jpoirier/gortlsdr/releases)
[![Build Status](http://circleci-badges-max.herokuapp.com/img/jpoirier/gortlsdr/master?token=:circle-ci-token)](https://circleci.com/gh/jpoirier/gortlsdr/tree/master)
[![Godoc reference](https://godoc.org/github.com/jpoirier/gortlsdr?status.svg)](https://godoc.org/github.com/jpoirier/gortlsdr)
[![Go Report Card](http://goreportcard.com/badge/jpoirier/gortlsdr)](http://goreportcard.com/report/jpoirier/gortlsdr)
[![BSD3 License](http://img.shields.io/badge/license-BSD3-brightgreen.svg)](https://tldrlegal.com/license/bsd-3-clause-license-%28revised%29)

# Description

gortlsdr is a simple Go interface to devices supported by the RTL-SDR project, which turns certain USB DVB-T dongles
employing the Realtek RTL2832U chipset into a low-cost, general purpose software-defined radio receiver. It wraps all
the functions in the [librtlsdr library](http://sdr.osmocom.org/trac/wiki/rtl-sdr) (including asynchronous read support).

Supported Platforms:
* Linux
* OS X
* Windows (with a little work)


# Installation

## Dependencies
* [Go tools](https://golang.org)
* [librtlsdr] (http://sdr.osmocom.org/trac/wiki/rtl-sdr) - builds dated after 5/5/12
* [libusb] (https://www.libusb.org)
* [git] (https://git-scm.com)


## Usage
All functions in librtlsdr are accessible from the gortlsdr package:

    go get -u github.com/jpoirier/gortlsdr

## Example
See the rtlsdr_eample.go file:

    go run rtlsdr_example.go

## Windows
If you don't want to build the librtlsdr and libusb dependencies from source you can use the librtlsdr pre-built package,
which includes libusb, but you're restricted to building a 32-bit gortlsdr library.

Building gortlsdr on Windows:
* Download and install [git](http://git-scm.com).
* Download and install the [Go tools](https://code.google.com/p/go/downloads/list?q=OpSys-Windows+Type%3DInstaller).
  Create a "go-pkgs" directory-your user folder is a good location-and add a GOPATH variable to your system environment, where
  GOPATH is set to the go-pkgs path, e.g. GOPATH=c:\users\jpoirier\go-pkgs.
* Download the pre-built [rtl-sdr library](http://sdr.osmocom.org/trac/attachment/wiki/rtl-sdr/RelWithDebInfo.zip) and unzip
  it, e.g. to your user folder. Note the path to the header files and the *.dll files in the x32 folder.
* Download gortlsdr, but don't install the package:

          go get -d github.com/jpoirier/gortlsdr

* Set CFLAGS and LDFLAGS in rtlsdr.go. Open the rtlsdr.go file in an editor, it'll be in go-pkgs\src\github.com\jpoirier\gortlsdr,
  and set the following two windows specific flags shown below, but with the correct paths from your system. CFLAGS points to
  the header files and LDFLAGS to the *.dll files:

          cgo windows CFLAGS: -IC:/Users/jpoirier/rtlsdr
          cgo windows LDFLAGS: -lrtlsdr -LC:/Users/jpoirier/rtlsdr/x32

* Build gortlsdr:

          go install github.com/jpoirier/gortlsdr

* Insert the DVB-T/DAB/FM dongle into a USB port, open a shell window in go-pkgs\src\github.com\jpoirier\gortlsdr and run
  the example program: go run rtlsdr_example.go. Note, the pre-built rtl-sdr package contains several test executables as well.


# Credit
* [pyrtlsdr](https://github.com/roger-/pyrtlsdr) for the great read-me description, which I copied.
* [osmoconSDR] (http://sdr.osmocom.org/trac/wiki/rtl-sdr) for the rtl-sdr library.
* [Antti Palosaari] (http://thread.gmane.org/gmane.linux.drivers.video-input-infrastructure/44461/focus=44461) for sharing!

# Todo
* create a plotting example using [plotinum](https://code.google.com/p/plotinum)
* create a Go port of librtlsdr's rtl_fm.c
* remove the rtl-sdr dependency using [gousb](https://github.com/kylelemons/gousb) and [go-dsp](https://github.com/mjibson/go-dsp)

-joe
