// +build ignore

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	rtl "github.com/jpoirier/gortlsdr"
)

// UAT holds a device context.
type UAT struct {
	dev *rtl.Context
	wg  *sync.WaitGroup
}

// read does synchronous specific reads.
func (u *UAT) read() {
	defer u.wg.Done()
	log.Println("Entered UAT read() ...")

	var readCnt uint64
	var buffer = make([]uint8, rtl.DefaultBufLength)
	for {
		nRead, err := u.dev.ReadSync(buffer, rtl.DefaultBufLength)
		if err != nil {
			// log.Printf("\tReadSync Failed - error: %s\n", err)
			break
		}
		// log.Printf("\tReadSync %d\n", nRead)
		if nRead > 0 {
			// buf := buffer[:nRead]
			fmt.Printf("\rnRead %d: readCnt: %d", nRead, readCnt)
			readCnt++
		}
	}
}

// shutdown
func (u *UAT) shutdown() {
	fmt.Println("\n")
	log.Println("\nEntered UAT shutdown() ...")
	log.Println("UAT shutdown(): closing device ...")
	u.dev.Close() // preempt the blocking ReadSync call
	log.Println("UAT shutdown(): calling uatWG.Wait() ...")
	u.wg.Wait() // Wait for the goroutine to shutdown
	log.Println("UAT shutdown(): uatWG.Wait() returned...")
}

// sdrConfig configures the device to 978 MHz UAT channel.
func (u *UAT) sdrConfig(indexID int) (err error) {
	if u.dev, err = rtl.Open(indexID); err != nil {
		log.Printf("\tUAT Open Failed...\n")
		return
	}
	log.Printf("\tGetTunerType: %s\n", u.dev.GetTunerType())

	//---------- Set Tuner Gain ----------
	err = u.dev.SetTunerGainMode(true)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetTunerGainMode Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetTunerGainMode Successful\n")

	tgain := 480
	err = u.dev.SetTunerGain(tgain)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetTunerGain Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetTunerGain Successful\n")

	//---------- Get/Set Sample Rate ----------
	samplerate := 2083334
	err = u.dev.SetSampleRate(samplerate)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetSampleRate Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetSampleRate - rate: %d\n", samplerate)
	log.Printf("\tGetSampleRate: %d\n", u.dev.GetSampleRate())

	//---------- Get/Set Xtal Freq ----------
	rtlFreq, tunerFreq, err := u.dev.GetXtalFreq()
	if err != nil {
		u.dev.Close()
		log.Printf("\tGetXtalFreq Failed - error: %s\n", err)
		return
	}
	log.Printf("\tGetXtalFreq - Rtl: %d, Tuner: %d\n", rtlFreq, tunerFreq)

	newRTLFreq := 28800000
	newTunerFreq := 28800000
	err = u.dev.SetXtalFreq(newRTLFreq, newTunerFreq)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetXtalFreq Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetXtalFreq - Center freq: %d, Tuner freq: %d\n",
		newRTLFreq, newTunerFreq)

	//---------- Get/Set Center Freq ----------
	err = u.dev.SetCenterFreq(978000000)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetCenterFreq 978MHz Failed, error: %s\n", err)
		return
	}
	log.Printf("\tSetCenterFreq 978MHz Successful\n")

	log.Printf("\tGetCenterFreq: %d\n", u.dev.GetCenterFreq())

	//---------- Set Bandwidth ----------
	bw := 1000000
	log.Printf("\tSetting Bandwidth: %d\n", bw)
	if err = u.dev.SetTunerBw(bw); err != nil {
		u.dev.Close()
		log.Printf("\tSetTunerBw %d Failed, error: %s\n", bw, err)
		return
	}
	log.Printf("\tSetTunerBw %d Successful\n", bw)

	if err = u.dev.ResetBuffer(); err != nil {
		u.dev.Close()
		log.Printf("\tResetBuffer Failed - error: %s\n", err)
		return
	}
	log.Printf("\tResetBuffer Successful\n")

	//---------- Get/Set Freq Correction ----------
	freqCorr := u.dev.GetFreqCorrection()
	log.Printf("\tGetFreqCorrection: %d\n", freqCorr)
	err = u.dev.SetFreqCorrection(freqCorr)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetFreqCorrection %d Failed, error: %s\n", freqCorr, err)
		return
	}
	log.Printf("\tSetFreqCorrection %d Successful\n", freqCorr)

	return
}

// sigAbort
func (u *UAT) sigAbort() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	u.shutdown()
	os.Exit(0)
}

func main() {
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
	indexID := 0
	log.Printf("===== Device name: %s =====\n", rtl.GetDeviceName(indexID))
	log.Printf("===== Running tests using device indx: 0 =====\n")

	uatDev := &UAT{}
	if err := uatDev.sdrConfig(indexID); err != nil {
		log.Fatal("uatDev = &UAT{indexID: id} failed: %s\n", err)
	}
	uatDev.wg = &sync.WaitGroup{}
	uatDev.wg.Add(1)
	fmt.Printf("\n======= CTRL+C to exit... =======\n\n")
	go uatDev.read()
	uatDev.sigAbort()
}
