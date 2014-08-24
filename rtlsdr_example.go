// +build ignore

package main

import (
	rtl "github.com/jpoirier/gortlsdr"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	// "unsafe"
)

func rtlsdr_cb(buf []byte, userctx *rtl.UserCtx) {
	log.Printf("Length of async-read buffer: %d", len(buf))
	if c, ok := (*userctx).(chan bool); ok {
		c <- true // async-read done signal
	}
}

func async_stop(dev *rtl.Context, c chan bool) {
	<-c // async-read done signal

	log.Println("Received async-read done, calling CancelAsync")
	if status := dev.CancelAsync(); status != rtl.Success {
		log.Println("CancelAsync failed")
	} else {
		log.Println("CancelAsync successful")
	}

	os.Exit(0)
}

func sig_abort(dev *rtl.Context) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	_ = dev.CancelAsync()
	dev.Close()
	os.Exit(0)
}

func main() {
	runtime.GOMAXPROCS(3)
	var status int
	var dev *rtl.Context

	if c := rtl.GetDeviceCount(); c == 0 {
		log.Fatal("No devices found, exiting.\n")
	} else {
		for i := 0; i < c; i++ {
			m, p, s, status := rtl.GetDeviceUsbStrings(i)
			log.Printf("GetDeviceUsbStrings %s - %s %s %s\n",
				rtl.Status[status], m, p, s)
		}
	}

	log.Printf("===== Device name: %s =====\n", rtl.GetDeviceName(0))
	log.Printf("===== Running tests using device indx: %d =====\n", 0)

	if dev, status = rtl.Open(0); status != rtl.Success {
		log.Fatal("\tOpen Failed, exiting\n")
	}
	defer dev.Close()
	go sig_abort(dev)

	m, p, s, status := dev.GetUsbStrings()
	log.Printf("\tGetUsbStrings %s - %s %s %s\n", rtl.Status[status], m, p, s)

	g, status := dev.GetTunerGains()
	log.Printf("\tGetTunerGains %s\n", rtl.Status[status])
	if status == rtl.Success {
		log.Printf("\tGains: ")
		for _, j := range g {
			log.Printf("%d ", j)
		}
		log.Printf("\n")
	}

	log.Printf("\tSetSampleRate %s - rate: %d\n",
		rtl.Status[dev.SetSampleRate(rtl.DefaultSampleRate)], rtl.DefaultSampleRate)
	log.Printf("\tGetSampleRate: %d\n", dev.GetSampleRate())

	// status = dev.SetXtalFreq(rtl_freq, tuner_freq)
	// log.Printf("\tSetXtalFreq %s - Center freq: %d, Tuner freq: %d\n",
	// 	rtl.Status[status], rtl_freq, tuner_freq)

	rtl_freq, tuner_freq, status := dev.GetXtalFreq()
	log.Printf("\tGetXtalFreq %s - Rtl: %d, Tuner: %d\n",
		rtl.Status[status], rtl_freq, tuner_freq)

	status = dev.SetCenterFreq(850000000)
	if status < 0 {
		log.Printf("\tSetCenterFreq 850MHz Failed, error code: %d\n", status)
	} else {
		log.Printf("\tSetCenterFreq 850MHz Successful\n")
	}

	log.Printf("\tGetCenterFreq: %d\n", dev.GetCenterFreq())
	log.Printf("\tGetFreqCorrection: %d\n", dev.GetFreqCorrection())
	log.Printf("\tGetTunerType: %s\n", rtl.TunerType[dev.GetTunerType()])
	log.Printf("\tSetTunerGainMode: %s\n", rtl.TunerType[dev.SetTunerGainMode(rtl.GainAuto)])
	log.Printf("\tGetTunerGain: %d\n", dev.GetTunerGain())

	/*
		func (c *Context) SetFreqCorrection(ppm int) (err int)
		func (c *Context) SetTunerGain(gain int) (err int)
		func (c *Context) SetTunerIfGain(stage, gain int) (err int)
		func (c *Context) SetAgcMode(on int) (err int)
		func (c *Context) SetDirectSampling(on int) (err int)
	*/

	if status = dev.SetTestMode(1); status == 0 {
		log.Printf("\tSetTestMode 'On' Successful\n")
	} else {
		log.Printf("\tSetTestMode 'On' Failed - error code: %d\n", status)
	}

	log.Printf("\tResetBuffer %s\n", rtl.Status[dev.ResetBuffer()])

	var buffer []byte = make([]uint8, rtl.DefaultBufLength)
	n_read, status := dev.ReadSync(buffer, rtl.DefaultBufLength)
	log.Printf("\tReadSync %s\n", rtl.Status[status])
	if status == rtl.Success && n_read < rtl.DefaultBufLength {
		log.Printf("ReadSync short read, %d samples lost\n", rtl.DefaultBufLength-n_read)
	}

	if status = dev.SetTestMode(1); status == 0 {
		log.Printf("\tSetTestMode 'Off' Successful\n")
	} else {
		log.Printf("\tSetTestMode 'Off' Fail - error code: %d\n", status)
	}

	// Note, ReadAsync blocks until CancelAsync is called, so spawn
	// a goroutine running in its own system thread that'll wait
	// for the async-read callback to signal when it's done.
	IQch := make(chan bool)
	go async_stop(dev, IQch)
	var userctx rtl.UserCtx = IQch
	status = dev.ReadAsync(rtlsdr_cb, &userctx, rtl.DefaultAsyncBufNumber, rtl.DefaultBufLength)
	log.Printf("\tReadAsync %s\n", rtl.Status[status])

	log.Printf("Exiting...\n")
}
