// +build ignore

package main

import (
	rtl "github.com/jpoirier/gortlsdr"
	"log"
	"runtime"
	// "unsafe"
)

// TODO: pass the channel via the callback UserCtx
var c1 = make(chan bool)

func rtlsdr_cb(buf []int8, userctx *rtl.UserCtx) {
	log.Printf("Length of buffer: %d", len(buf))
	c1 <- true // tell main we're done
}

func async_stop(dev *rtl.Context) {
	<-c1 // wait for a signal

	log.Println("Received signal form callback, calling CancelAsync\n")
	if ok := dev.CancelAsync(); ok != rtl.Success {
		log.Fatal("ReadSync failed\n")
	}
}

func main() {
	runtime.GOMAXPROCS(3)

	if c := rtl.GetDeviceCount(); c == 0 {
		log.Fatal("No devices found.\n")
	} else {
		for i := 0; i < c; i++ {
			m, p, s, err := rtl.GetDeviceUsbStrings(i)
			log.Printf("Device USB Striing - err: %d, m: %s, p: %s, s: %s\n", err, m, p, s)
		}
	}

	log.Printf("Device name: %s\n", rtl.GetDeviceName(0))

	log.Printf("Using device indx %d\n", 0)
	var ok int
	var dev *rtl.Context
	if dev, ok = rtl.Open(0); ok != rtl.Success {
		log.Fatal("Failed to open the device\n")
	}
	defer dev.Close()

	if m, p, s, ok := dev.GetUsbStrings(); ok != rtl.Success {
		log.Fatal("GetUsbStrings failed, exiting\n")
	} else {
		log.Printf("USB strings - m: %s, p: %s, s: %s\n", m, p, s)
	}

	if g, ok := dev.GetTunerGains(); ok != rtl.Success {
		log.Fatal("GetTunerGains failed, exiting\n")
	} else {
		for i, j := range g {
			log.Printf("Gain %d: %d\n", i, j)
		}
	}

	if rate, ok := dev.GetSampleRate(); ok == rtl.Error {
		// rtl-sdr lib needs fixing/
		log.Printf("GetSampleRate: %d\n", rate)
		// log.Fatal("GetCenterFreq failed, exiting\n")
	} else {
		log.Printf("GetSampleRate: %d\n", rate)
	}

	log.Printf("Setting sample rate to %d\n", rtl.DefaultSampleRate)
	if ok = dev.SetSampleRate(rtl.DefaultSampleRate); ok != rtl.Success {
		log.Fatal("SetSampleRate failed, exiting\n")
	}

	var rtl_freq, tuner_freq int
	if rtl_freq, tuner_freq, ok = dev.GetXtalFreq(); ok != rtl.Success {
		log.Fatal("GetXtalFreq failed, exiting\n")
	} else {
		log.Printf("GetXtalFreq - Center freq: %d, Tuner freq: %d\n", rtl_freq, tuner_freq)
	}

	if ok = dev.SetXtalFreq(rtl_freq, tuner_freq); ok != rtl.Success {
		log.Fatal("SetXtalFreq failed, exiting\n")
	}

	if freq, ok := dev.GetCenterFreq(); ok == rtl.Error {
		// rtl-sdr lib needs fixing/
		log.Printf("Center freq: %d\n", freq)
		// log.Fatal("GetCenterFreq failed, exiting\n")
	} else {
		log.Printf("Center freq: %d\n", freq)
	}

	// ok = dev.SetCenterFreq(freq)
	// if ok < 0 {
	// 	log.Printf("Error code: %d\n", ok)
	// 	log.Fatal("SetCenterFreq failed, exiting\n")
	// } else {
	// 	log.Printf("Center freq set: %d\n", freq)
	// }

	if freq, ok := dev.GetFreqCorrection(); ok != rtl.Success {
		// rtl-sdr lib needs fixing/
		log.Printf("GetFreqCorrection: %d\n", freq)
		// log.Fatal("GetCenterFreq failed, exiting\n")
	} else {
		log.Printf("GetFreqCorrection: %d\n", freq)
	}

	rtlsdr_tuner := dev.GetTunerType()
	log.Printf("GetTunerType: %s\n", rtl.TypeMap[rtlsdr_tuner])

	/*

		func (c *Context) SetFreqCorrection(ppm int) (err int)
		func (c *Context) SetTunerGain(gain int) (err int)
		func (c *Context) SetTunerIfGain(stage, gain int) (err int)
		func (c *Context) SetTunerGainMode(manual int)
		func (c *Context) SetAgcMode(on int) (err int)
		func (c *Context) SetDirectSampling(on int) (err int)
	*/

	if ok = dev.SetTestMode(1); ok < 1 {
		log.Printf("SetTestMode to on failed with error code: %d\n", ok)
		log.Fatal("")
	}

	if ok = dev.ResetBuffer(); ok != rtl.Success {
		log.Fatal("Buffer reset failed, exiting\n")
	}

	var buffer []byte = make([]uint8, rtl.DefaultBufLength)
	if n_read, ok := dev.ReadSync(buffer, rtl.DefaultBufLength); ok != rtl.Success {
		log.Fatal("ReadSync failed, exiting\n")
	} else {
		if n_read < rtl.DefaultBufLength {
			log.Fatal("ReadSync short read, samples lost, exiting\n")
		}
	}

	log.Println("ReadSync successful")
	// log.Println(buffer)

	if ok = dev.SetTestMode(0); ok < 1 {
		log.Printf("SetTestMode to off failed with error code: %d\n", ok)
		log.Fatal("")
	}
	/* 	Unusable on my system (OSX 10.7.5) due to a segfault caused by
	a call to libusb_handle_events_timeout in the libusb library
	from librtlsdr - it's a known issue.

	// ReadAsync blocks until CancelAsync is called, so spawn
	// a goroutine, running in a system thread, that waits
	// for a signal from the callback function that it's
	// done.
	log.Println("Calling ReadAsync")
	go async_stop()
	var userctx rtl.UserCtx
	ok = dev.ReadAsync(rtlsdr_cb, &userctx, rtl.DefaultAsyncBufNumber, rtl.DefaultBufLength)
	if ok != rtl.Success {
		log.Fatal("ReadAsync failed, exiting\n")
	}
	*/
	log.Printf("Closing...\n")
}
