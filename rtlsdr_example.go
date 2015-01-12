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

var send_ping bool = true

// rtlsdr_cb is used for asynchronous streaming. It's the
// user callback function passed to librtlsdr.
func rtlsdr_cb(buf []byte, userctx *rtl.UserCtx) {
	if send_ping {
		send_ping = false
		// send a ping to async_stop
		if c, ok := (*userctx).(chan bool); ok {
			c <- true // async-read done signal
		}
	}
	log.Printf("Length of async-read buffer: %d\n", len(buf))
}

// async_stop pends for a ping from the rtlsdrCb function
// callback, and when received cancel the async callback.
func async_stop(dev *rtl.Context, c chan bool) {
	log.Println("async_stop running...")
	<-c
	log.Println("Received ping from rtlsdr_cb, calling CancelAsync")
	if err := dev.CancelAsync(); err != nil {
		log.Printf("CancelAsync failed - %s\n", err)
	} else {
		log.Printf("CancelAsync successful\n")
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

	//---------- Device Check ----------
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
	log.Printf("===== Running tests using device indx: 0 =====\n")

	//---------- Open Device ----------
	if dev, err = rtl.Open(0); err != nil {
		log.Fatal("\tOpen Failed, exiting\n")
	}
	defer dev.Close()
	go sig_abort(dev)

	//---------- Device Strings ----------
	m, p, s, err := dev.GetUsbStrings()
	if err != nil {
		log.Printf("\tGetUsbStrings Failed - error: %s\n", err)
	} else {
		log.Printf("\tGetUsbStrings - %s %s %s\n", m, p, s)
	}

	log.Printf("\tGetTunerType: %s\n", dev.GetTunerType())

	//---------- Get/Set Tuner Gains ----------
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

	tgain := dev.GetTunerGain()
	log.Printf("\tGetTunerGain: %d\n", tgain)

	err = dev.SetTunerGainMode(true)
	if err != nil {
		log.Printf("\tSetTunerGainMode Failed - error: %s\n", err)
	} else {
		log.Printf("\tSetTunerGainMode Successful\n")
	}

	err = dev.SetTunerGain(tgain)
	if err != nil {
		log.Printf("\tSetTunerGain Failed - error: %s\n", err)
	} else {
		log.Printf("\tSetTunerGain Successful\n")
	}

	//---------- Get/Set Sample Rate ----------
	err = dev.SetSampleRate(rtl.DefaultSampleRate)
	if err != nil {
		log.Printf("\tSetSampleRate Failed - error: %s\n", err)
	} else {
		log.Printf("\tSetSampleRate - rate: %d\n", rtl.DefaultSampleRate)
	}
	log.Printf("\tGetSampleRate: %d\n", dev.GetSampleRate())

	//---------- Get/Set Xtal Freq ----------
	rtl_freq, tuner_freq, err := dev.GetXtalFreq()
	if err != nil {
		log.Printf("\tGetXtalFreq Failed - error: %s\n", err)
	} else {
		log.Printf("\tGetXtalFreq - Rtl: %d, Tuner: %d\n", rtl_freq, tuner_freq)
	}

	err = dev.SetXtalFreq(rtl_freq, tuner_freq)
	if err != nil {
		log.Printf("\tSetXtalFreq Failed - error: %s\n", err)
	} else {
		log.Printf("\tSetXtalFreq - Center freq: %d, Tuner freq: %d\n",
			rtl_freq, tuner_freq)
	}

	//---------- Get/Set Center Freq ----------
	err = dev.SetCenterFreq(850000000)
	if err != nil {
		log.Printf("\tSetCenterFreq 850MHz Failed, error: %s\n", err)
	} else {
		log.Printf("\tSetCenterFreq 850MHz Successful\n")
	}

	log.Printf("\tGetCenterFreq: %d\n", dev.GetCenterFreq())

	//---------- Get/Set Freq Correction ----------
	freq_corr := dev.GetFreqCorrection()
	log.Printf("\tGetFreqCorrection: %d\n", freq_corr)
	err = dev.SetFreqCorrection(freq_corr)
	if err != nil {
		log.Printf("\tGetFreqCorrection Failed, error: %s\n", err)
	} else {
		log.Printf("\tGetFreqCorrection Successful\n")
	}

	//---------- Get/Set AGC Mode ----------
	if err = dev.SetAgcMode(false); err == nil {
		log.Printf("\tSetAgcMode off Successful\n")
	} else {
		log.Printf("\tSetAgcMode Failed, error: %s\n", err)
	}

	//---------- Get/Set Direct Sampling ----------
	if mode, err := dev.GetDirectSampling(); err == nil {
		log.Printf("\tGetDirectSampling Successful, mode: %s\n",
			rtl.SamplingModes[mode])
	} else {
		log.Printf("\tSetTestMode 'On' Failed - error: %s\n", err)
	}

	if err = dev.SetDirectSampling(rtl.SamplingNone); err == nil {
		log.Printf("\tSetDirectSampling 'On' Successful\n")
	} else {
		log.Printf("\tSetDirectSampling 'On' Failed - error: %s\n", err)
	}

	//---------- Get/Set Tuner IF Gain ----------
	// if err = SetTunerIfGain(stage, gain: int); err == nil {
	// 	log.Printf("\SetTunerIfGain Successful\n")
	// } else {
	// 	log.Printf("\tSetTunerIfGain Failed - error: %s\n", err)
	// }

	//---------- Get/Set test mode ----------
	if err = dev.SetTestMode(true); err == nil {
		log.Printf("\tSetTestMode 'On' Successful\n")
	} else {
		log.Printf("\tSetTestMode 'On' Failed - error: %s\n", err)
	}

	if err = dev.SetTestMode(false); err == nil {
		log.Printf("\tSetTestMode 'Off' Successful\n")
	} else {
		log.Printf("\tSetTestMode 'Off' Fail - error: %s\n", err)
	}

	//---------- Get/Set misc. streaming ----------
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
