// +build ignore

package main

import (
	rtl "github.com/jpoirier/gortlsdr"
	"log"
	"runtime"
	// "unsafe"
)

// TODO: pass chan to the callback function as a *UserCtx
var c1 = make(chan bool)

func rtlsdr_cb(buf []int8, userctx *rtl.UserCtx) {
	log.Printf("Length of async-read buffer: %d", len(buf))
	c1 <- true // async-read done signal
}

func async_stop(dev *rtl.Context) {
	<-c1 // async-read done signal

	log.Println("Received async-read done, calling CancelAsync\n")
	if status := dev.CancelAsync(); status != rtl.Success {
		log.Println("CancelAsync failed\n")
	} else {
		log.Println("CancelAsync successful\n")
	}
}

func main() {
	runtime.GOMAXPROCS(2)
	var status int
	var dev *rtl.Context

	if c := rtl.GetDeviceCount(); c == 0 {
		log.Fatal("No devices found, exiting.\n")
	} else {
		for i := 0; i < c; i++ {
			m, p, s, status := rtl.GetDeviceUsbStrings(i)
			log.Printf("GetDeviceUsbStrings %s: %d, m: %s, p: %s, s: %s\n",
				rtl.Status[status], m, p, s)
		}
	}

	log.Printf("===== Device name: %s =====\n", rtl.GetDeviceName(0))
	log.Printf("===== Running tests using device indx: %d =====\n", 0)

	if dev, status = rtl.Open(0); status != rtl.Success {
		log.Fatal("\tOpen Failed, exiting\n")
	}
	defer dev.Close()

	m, p, s, status := dev.GetUsbStrings()
	log.Printf("\tGetUsbStrings %s - m: %s, p: %s, s: %s\n", rtl.Status[status], m, p, s)

	g, status := dev.GetTunerGains()
	log.Printf("\tGetTunerGains %s\n", rtl.Status[status])
	if status == rtl.Success {
		for i, j := range g {
			log.Printf("\t\tGain %d: %d\n", i, j)
		}
	}

	rate, status := dev.GetSampleRate()
	log.Printf("\tGetSampleRate %s - rate: %d\n", rtl.Status[status], rate)

	status = dev.SetSampleRate(rtl.DefaultSampleRate)
	log.Printf("\tSetSampleRate %s - rate: %d\n", rtl.Status[status], rtl.DefaultSampleRate)

	rtl_freq, tuner_freq, status := dev.GetXtalFreq()
	log.Printf("\tGetXtalFreq %s - Center freq: %d, Tuner freq: %d\n",
		rtl.Status[status], rtl_freq, tuner_freq)

	status = dev.SetXtalFreq(rtl_freq, tuner_freq)
	log.Printf("\tSetXtalFreq %s - Center freq: %d, Tuner freq: %d\n",
		rtl.Status[status], rtl_freq, tuner_freq)

	freq, status := dev.GetCenterFreq()
	log.Printf("\tGetCenterFreq %s - freq: %d\n", rtl.Status[status], freq)

	// status = dev.SetCenterFreq(freq)
	// if status < 0 {
	// 	log.Printf("Error code: %d\n", status)
	// 	log.Println("SetCenterFreq failed\n")
	// } else {
	// 	log.Printf("Center freq set: %d\n", freq)
	// }

	freq, status = dev.GetFreqCorrection()
	log.Printf("\tGetFreqCorrection %s - freq: %d\n", rtl.Status[status], freq)

	rtlsdr_tuner := dev.GetTunerType()
	log.Printf("\tGetTunerType %s - tuner type: %d\n", rtl.Status[status], rtl.TunerType[rtlsdr_tuner])

	/*

		func (c *Context) SetFreqCorrection(ppm int) (err int)
		func (c *Context) SetTunerGain(gain int) (err int)
		func (c *Context) SetTunerIfGain(stage, gain int) (err int)
		func (c *Context) SetTunerGainMode(manual int)
		func (c *Context) SetAgcMode(on int) (err int)
		func (c *Context) SetDirectSampling(on int) (err int)
	*/

	if status = dev.SetTestMode(1); status < 1 {
		log.Printf("\tSetTestMode '1' Fail - error code: %d\n", status)
	} else {
		log.Printf("\tSetTestMode Success\n")
	}

	status = dev.ResetBuffer()
	log.Printf("\tResetBuffer %s\n", rtl.Status[status])

	var buffer []byte = make([]uint8, rtl.DefaultBufLength)
	n_read, status := dev.ReadSync(buffer, rtl.DefaultBufLength)
	log.Printf("\tReadSync %s\n", rtl.Status[status])
	if status == rtl.Success {
		if n_read < rtl.DefaultBufLength {
			log.Println("ReadSync short read, %d samples lost\n", rtl.DefaultBufLength-n_read)
		}
	}

	if status = dev.SetTestMode(1); status < 1 {
		log.Printf("\tSetTestMode '0' Fail - error code: %d\n", status)
	} else {
		log.Printf("\tSetTestMode '0' Success\n")
	}

	/*
		Calling ReadAsync on my systems, OSX 10.7.5 and 64-bit Xubuntu, fails due
		to a segfault that seems to manifest in libusb's libusb_handle_events_timeout
		function, which is a known issue.
	*/
	// Note, ReadAsync blocks until CancelAsync is called, so spawn
	// a goroutine running in its own system thread that'll wait
	// for the async-read callback to signal when it's done.
	go async_stop(dev)
	var userctx rtl.UserCtx
	status = dev.ReadAsync(rtlsdr_cb, &userctx, rtl.DefaultAsyncBufNumber, rtl.DefaultBufLength)
	log.Printf("\tReadAsync %s\n", rtl.Status[status])

	log.Printf("Exiting...\n")
}
