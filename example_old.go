// +build ignore

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	rtl "github.com/jpoirier/gortlsdr"
	// "unsafe"
)

var sendPing = true

//IQch async-read done signal
var IQch = make(chan bool)

// rtlsdrCb is used for asynchronous streaming. It's our
// callback function rtl-sdr callback function.
func rtlsdrCb(buf []byte) {
	if sendPing {
		log.Println("sendPing = false")
		sendPing = false
		IQch <- true // async-read done signal
	}
	log.Printf("Length of async-read buffer: %d\n", len(buf))
}

// asyncStop pends for a ping from the rtlsdrCb function
// callback, and when received cancel the async callback.
func asyncStop(dev *rtl.Context, c chan bool) {
	log.Println("asyncStop running...")
	<-c
	log.Println("Received ping from rtlsdrCb, calling CancelAsync")
	if err := dev.CancelAsync(); err != nil {
		log.Printf("CancelAsync failed - %s\n", err)
	} else {
		log.Printf("CancelAsync successful\n")
	}
	//os.Exit(0)
}

func sigAbort(dev *rtl.Context) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	_ = dev.CancelAsync()
	dev.Close()
	os.Exit(0)
}

func main() {
	var err error
	var dev *rtl.Context

	//---------- Device Check ----------
	if c := rtl.GetDeviceCount(); c == 0 {
		log.Fatal("No devices found, exiting.\n")
	} else {
		for i := 0; i < c; i++ {
			m, p, s, err := rtl.GetDeviceUsbStrings(i)
			if err == nil {
				err = errors.New("")
			}
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
	go sigAbort(dev)

	//---------- Device Strings ----------
	m, p, s, err := dev.GetUsbStrings()
	if err != nil {
		log.Printf("\tGetUsbStrings Failed - error: %s\n", err)
	} else {
		log.Printf("\tGetUsbStrings - %s %s %s\n", m, p, s)
	}

	log.Printf("\tGetTunerType: %s\n", dev.GetTunerType())

	//---------- Device Info ----------
	info, err := dev.GetHwInfo()
	if err != nil {
		log.Printf("\tGetHwInfo Failed - error: %s\n", err)
	} else {
		log.Printf("\tVendor ID      : 0x%X\n", info.VendorID)
		log.Printf("\tProduct ID     : 0x%X\n", info.ProductID)
		log.Println("\tManufacturer  : ", info.Manufact)
		log.Println("\tProduct       : ", info.Product)
		log.Println("\tSerial        : ", info.Serial)
		log.Println("\tHave Serial   : ", info.HaveSerial)
		log.Println("\tEnable IR     : ", info.EnableIR)
		log.Println("\tRemote Wakeup : ", info.RemoteWakeup)

		// err = dev.SetHwInfo(info)
		// if err != nil {
		// 	log.Printf("\tSetHwInfo Failed - error: %s\n", err)
		// } else {
		// 	log.Printf("\ttSetHwInfo Successful\n")
		// }
	}

	//---------- Get/Set Tuner Gains ----------
	tgain := dev.GetTunerGain()
	log.Printf("\tGetTunerGain: %d\n", tgain)

	g, err := dev.GetTunerGains()
	if err != nil {
		log.Printf("\tGetTunerGains Failed - error: %s\n", err)
	} else if len(g) > 0 {
		gains := fmt.Sprintf("\tGains: ")
		for _, j := range g {
			gains += fmt.Sprintf("%d ", j)
		}
		log.Printf("%s\n", gains)
		tgain = g[0]
	}

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
	rtlFreq, tunerFreq, err := dev.GetXtalFreq()
	if err != nil {
		log.Printf("\tGetXtalFreq Failed - error: %s\n", err)
	} else {
		log.Printf("\tGetXtalFreq - Rtl: %d, Tuner: %d\n", rtlFreq, tunerFreq)
	}

	err = dev.SetXtalFreq(rtlFreq, tunerFreq)
	if err != nil {
		log.Printf("\tSetXtalFreq Failed - error: %s\n", err)
	} else {
		log.Printf("\tSetXtalFreq - Center freq: %d, Tuner freq: %d\n",
			rtlFreq, tunerFreq)
	}

	//---------- Get/Set Center Freq ----------
	err = dev.SetCenterFreq(978000000)
	if err != nil {
		log.Printf("\tSetCenterFreq 978MHz Failed, error: %s\n", err)
	} else {
		log.Printf("\tSetCenterFreq 978MHz Successful\n")
	}

	log.Printf("\tGetCenterFreq: %d\n", dev.GetCenterFreq())

	//---------- Set Bandwidth ----------
	bandwidths := []int{300000, 400000, 550000, 700000, 1000000, 1200000,
		1300000, 1600000, 2200000, 3000000, 4000000, 5000000, 6000000, 7000000}
	for _, bw := range bandwidths {
		log.Printf("\tSetting Bandwidth: %d\n", bw)
		if err = dev.SetTunerBw(bw); err != nil {
			log.Printf("\tSetTunerBw %d Failed, error: %s\n", bw, err)
		} else {
			log.Printf("\tSetTunerBw %d Successful\n", bw)
		}
		time.Sleep(1 * time.Second)
	}

	//---------- Get/Set Freq Correction ----------
	freqCorr := dev.GetFreqCorrection()
	log.Printf("\tGetFreqCorrection: %d\n", freqCorr)
	err = dev.SetFreqCorrection(10) // 10ppm
	if err != nil {
		log.Printf("\tSetFreqCorrection %d Failed, error: %s\n", 10, err)
	} else {
		log.Printf("\tSetFreqCorrection %d Successful\n", 10)
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

	var buffer = make([]uint8, rtl.DefaultBufLength)
	nRead, err := dev.ReadSync(buffer, rtl.DefaultBufLength)
	if err != nil {
		log.Printf("\tReadSync Failed - error: %s\n", err)
	} else {
		log.Printf("\tReadSync %d\n", nRead)
	}
	if err == nil && nRead < rtl.DefaultBufLength {
		log.Printf("ReadSync short read, %d samples lost\n", rtl.DefaultBufLength-nRead)
	}

	// Note, ReadAsync blocks until CancelAsync is called, so spawn
	// a goroutine that waits for the async-read callback to signal it's done.
	go asyncStop(dev, IQch)

	err = dev.ReadAsync(rtlsdrCb, nil, rtl.DefaultAsyncBufNumber, rtl.DefaultBufLength)
	if err == nil {
		log.Printf("\tReadAsync Successful\n")
	} else {
		log.Printf("\tReadAsync Fail - error: %s\n", err)
	}
	log.Printf("Exiting...\n")
}
