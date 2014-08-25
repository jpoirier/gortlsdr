// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	rtl "github.com/jpoirier/gortlsdr"
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
	if err := dev.CancelAsync(); err != nil {
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
	var err error
	var dev *rtl.Context

	if c := rtl.GetDeviceCount(); c == 0 {
		log.Fatal("No devices found, exiting.\n")
	} else {
		for i := 0; i < c; i++ {
			m, p, s, err := rtl.GetDeviceUsbStrings(i)
			log.Printf("GetDeviceUsbStrings %s - %s %s %s\n",
				err, m, p, s)
		}
	}

	log.Printf("===== Device name: %s =====\n", rtl.GetDeviceName(0))
	log.Printf("===== Running tests using device indx: %d =====\n", 0)

	if dev, err = rtl.Open(0); err != nil {
		log.Fatal("\tOpen Failed, exiting\n")
	}
	defer dev.Close()
	go sig_abort(dev)

	m, p, s, err := dev.GetUsbStrings()
	if err != nil {
		log.Printf("\tGetUsbStrings Failed - error: %s\n", err)
	} else {
		log.Printf("\tGetUsbStrings - %s %s %s\n", m, p, s)
	}

	g, err := dev.GetTunerGains()
	if err != nil {
		log.Printf("\tGetTunerGains Failed - error: %s\n", err)
	} else {
		gains := fmt.Sprintf("\tGains: ")
		for _, j := range g {
			gains += fmt.Sprintf("%d ", j)
		}
		log.Printf("%s\n", gains)
	}

	err = dev.SetSampleRate(rtl.DefaultSampleRate)
	if err != nil {
		log.Printf("\tSetSampleRate Failed - error: %s\n", err)
	} else {
		log.Printf("\tSetSampleRate - rate: %d\n", rtl.DefaultSampleRate)
	}
	log.Printf("\tGetSampleRate: %d\n", dev.GetSampleRate())

	// status = dev.SetXtalFreq(rtl_freq, tuner_freq)
	// log.Printf("\tSetXtalFreq %s - Center freq: %d, Tuner freq: %d\n",
	// 	rtl.Status[status], rtl_freq, tuner_freq)

	rtl_freq, tuner_freq, err := dev.GetXtalFreq()
	if err != nil {
		log.Printf("\tGetXtalFreq Failed - error: %s\n", err)
	} else {
		log.Printf("\tGetXtalFreq - Rtl: %d, Tuner: %d\n", rtl_freq, tuner_freq)
	}

	err = dev.SetCenterFreq(850000000)
	if err != nil {
		log.Printf("\tSetCenterFreq 850MHz Failed, error: %s\n", err)
	} else {
		log.Printf("\tSetCenterFreq 850MHz Successful\n")
	}

	log.Printf("\tGetCenterFreq: %d\n", dev.GetCenterFreq())
	log.Printf("\tGetFreqCorrection: %d\n", dev.GetFreqCorrection())
	log.Printf("\tGetTunerType: %s\n", dev.GetTunerType())
	err = dev.SetTunerGainMode(false)
	if err != nil {
		log.Printf("\tSetTunerGainMode Failed - error: %s\n", err)
	} else {
		log.Printf("\tSetTunerGainMode Successful\n")
	}
	log.Printf("\tGetTunerGain: %d\n", dev.GetTunerGain())

	/*
		func (c *Context) SetFreqCorrection(ppm int) (err int)
		func (c *Context) SetTunerGain(gain int) (err int)
		func (c *Context) SetTunerIfGain(stage, gain int) (err int)
		func (c *Context) SetAgcMode(on int) (err int)
		func (c *Context) SetDirectSampling(on int) (err int)
	*/

	if err = dev.SetTestMode(true); err == nil {
		log.Printf("\tSetTestMode 'On' Successful\n")
	} else {
		log.Printf("\tSetTestMode 'On' Failed - error: %s\n", err)
	}

	if err = dev.ResetBuffer(); err == nil {
		log.Printf("\tResetBuffer Successful\n")
	} else {
		log.Printf("\tResetBuffer Failed - error: %s\n", err)
	}

	var buffer []byte = make([]uint8, rtl.DefaultBufLength)
	n_read, err := dev.ReadSync(buffer, rtl.DefaultBufLength)
	if err != nil {
		log.Printf("\tReadSync Failed - error: %s\n", err)
	} else {
		log.Printf("\tReadSync %d\n", n_read)
	}
	if err == nil && n_read < rtl.DefaultBufLength {
		log.Printf("ReadSync short read, %d samples lost\n", rtl.DefaultBufLength-n_read)
	}

	if err = dev.SetTestMode(false); err == nil {
		log.Printf("\tSetTestMode 'Off' Successful\n")
	} else {
		log.Printf("\tSetTestMode 'Off' Fail - error: %s\n", err)
	}

	// Note, ReadAsync blocks until CancelAsync is called, so spawn
	// a goroutine running in its own system thread that'll wait
	// for the async-read callback to signal when it's done.
	IQch := make(chan bool)
	go async_stop(dev, IQch)
	var userctx rtl.UserCtx = IQch
	err = dev.ReadAsync(rtlsdr_cb, &userctx, rtl.DefaultAsyncBufNumber, rtl.DefaultBufLength)
	if err == nil {
		log.Printf("\tReadAsync Successful\n")
	} else {
		log.Printf("\tReadAsync Fail - error: %s\n", err)
	}

	log.Printf("Exiting...\n")
}
