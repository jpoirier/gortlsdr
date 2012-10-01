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
- [libusb] (http://www.libusb.org)
- [git] (http://git-scm.com)


## Usage
All functions in librtlsdr are accessible in the gortlsdr package.

    go get -u github.com/jpoirier/gortlsdr

## Example
See the rtlsdr_eample.go file.

    go run rtlsdr_eample.go

## Windows
If you don't want to build the rtl-sdr dependency from source you can use the pre-built package, which includes libusb,
but you restricted you building a 32-bit gortlsdr library.

Building gortlsdr on Windows:
1. Download and install [git] (http://git-scm.com)
2. Download and install the [Go tools] (http://code.google.com/p/go/downloads/list?q=OpSys-Windows+Type%3DInstaller)
   Create a "go-pkgs" directory-your user folder is a good location-and add a GOPATH variable to your system environment, where
   GOPATH is set to the go-pkgs path, e.g. GOPATH=c:\users\jpoirier\go-pkgs.
3. Download the prebuilt [rtl-sdr library] (http://sdr.osmocom.org/trac/attachment/wiki/rtl-sdr/RelWithDebInfo.zip) and unzip
   it, e.g. to your user folder. Note the path to the header files and the *.dll files in the x32 folder.
4. Download gortlsdr but don't install the package, run the following command: go get -d github.com/jpoirier/gortlsdr
5. Set CFLAGS and LDFLAGS in rtlsdr.go. Open the rtlsdr.go file in an editor, it'll be in go-pkgs\src\github.com\jpoirier\gortlsdr,
   and set the following two windows specific flags shown below but with the correct paths on your system. CFLAGS points to
   the header files amd LDFLAGS to the *.dll files:

          #cgo windows CFLAGS: -IC:/Users/jpoirier/rtlsdr
          #cgo windows LDFLAGS: -lrtlsdr -LC:/Users/jpoirier/rtlsdr/x32
6. Build gortlsdr: go install github.com/jpoirier/gortlsdr
7. Insert the DVB-T/DAB/FM dongle into a USB port, open a shell window in go-pkgs\src\github.com\jpoirier\gortlsdr and run
   the exmaple program: go run rtlsdr_example.go

   The prebuilt rtl-sdr package also contains several test executables as well.


# Credit
The great read-me description was copied from the Python wrapper page [pyrtlsdr](https://github.com/roger-/pyrtlsdr/tree/master/rtlsdr).

-joe



