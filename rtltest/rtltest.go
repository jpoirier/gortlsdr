package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	rtl "github.com/jpoirier/gortlsdr"
)

func ConfigTest(dev *rtl.Context, samplerate, ppmErr int) error {

	//---------- Device Strings ----------
	m, p, s, err := dev.GetUsbStrings()
	if err != nil {
		log.Printf("\tGetUsbStrings Failed - error: %s\n", err)
		return err
	}
	log.Printf("\tGetUsbStrings - %s %s %s\n", m, p, s)

	log.Printf("\tGetTunerType: %s\n", dev.GetTunerType())

	//---------- Device Info ----------
	info, err := dev.GetHwInfo()
	if err != nil {
		log.Printf("\tGetHwInfo Failed - error: %s\n", err)
		return err
	}
	log.Printf("\tVendor ID      : 0x%X\n", info.VendorID)
	log.Printf("\tProduct ID     : 0x%X\n", info.ProductID)
	log.Println("\tManufacturer  : ", info.Manufact)
	log.Println("\tProduct       : ", info.Product)
	log.Println("\tSerial        : ", info.Serial)
	log.Println("\tHave Serial   : ", info.HaveSerial)
	log.Println("\tEnable IR     : ", info.EnableIR)
	log.Println("\tRemote Wakeup : ", info.RemoteWakeup)

	if ppmErr != 0 {
		//---------- Get/Set Freq Correction ----------
		//freqCorr := dev.GetFreqCorrection()
		log.Printf("\tGetFreqCorrection: %d\n", ppmErr)
		err = dev.SetFreqCorrection(ppmErr)
		if err != nil {
			log.Printf("\tSetFreqCorrection %d Failed, error: %s\n", ppmErr, err)
			return err
		}
		log.Printf("\tSetFreqCorrection %d Successful\n", ppmErr)
	}

	//---------- Get/Set Sample Rate ----------
	err = dev.SetSampleRate(samplerate)
	if err != nil {
		log.Printf("\tSetSampleRate Failed - error: %s\n", err)
		return err
	}
	samp_rate = samplerate
	log.Printf("\tSetSampleRate - rate: %d\n", samplerate)
	log.Printf("\tGetSampleRate: %d\n", dev.GetSampleRate())

	if err = dev.SetTestMode(true); err != nil {
		log.Printf("\tSetTestMode 'On' Failed - error: %s\n", err)
		return err
	}
	log.Printf("\tSetTestMode 'On' Successful\n")

	if err = dev.ResetBuffer(); err != nil {
		log.Printf("\tResetBuffer Failed - error: %s\n", err)
		return err
	}
	log.Printf("\tResetBuffer Successful\n")

	return nil
}

var (
	first        bool
	bcnt         uint8
	f            int
	glost, clost int
	samp_rate    int
)

func rtlsdrCb(buf []byte) {
	if !first {
		first = true
		bcnt = buf[0]
		//fmt.Println(bcnt, buf[0])
	}
	lost := 0

	for _, v := range buf {
		if bcnt != v {
			//fmt.Println(f, i, v, bcnt)
			if v > bcnt {
				lost += int(v - bcnt)
			} else {
				lost += 256 + int(bcnt-v)
			}

			clost++
			glost += lost
			bcnt = v
		}
		bcnt++
	}
	if lost > 0 {
		fmt.Printf("%v, lost at least %d bytes\n", f, lost)
	}
	f++
	// time.Sleep(time.Microsecond * 10)
	ppm_test(len(buf) / 2)
}

func main() {
	var device = flag.Int("d", 0, "device idx to use")
	var samplerate = flag.Int("s", 2048000, "Sample rate Hz")
	var ppmErr = flag.Int("p", 0, "PPM Error")
	var syncTest = flag.Bool("S", false, "use sync calls")
	flag.Parse()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)

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
	log.Printf("===== Device name: %s =====\n", rtl.GetDeviceName(*device))

	dev, err := rtl.Open(*device)
	if err != nil {
		return
	}
	defer dev.Close()

	err = ConfigTest(dev, *samplerate, *ppmErr)
	if err != nil {
		log.Fatalf("Config failed: %s\n", err.Error())
	}

	if *syncTest {
		var buffer = make([]uint8, rtl.DefaultBufLength)
	FOR:
		for {

			n, err := dev.ReadSync(buffer, rtl.DefaultBufLength)
			if err != nil {
				log.Fatal(err.Error())
				break FOR
			}
			rtlsdrCb(buffer[:n])
			dev.ReadSync(buffer, rtl.DefaultBufLength)
			rtlsdrCb(buffer)
			dev.ReadSync(buffer, rtl.DefaultBufLength)
			rtlsdrCb(buffer)

			select {
			case <-ch:
				break FOR
			default:
			}
		}
	} else {

		errch := make(chan error)
		// if in goroutine that call ReadAsync is call to defer dev.Close(),
		// in case of panic in callback, program deadlock.
		go func() {
			errch <- dev.ReadAsync(rtlsdrCb, nil, rtl.DefaultAsyncBufNumber, rtl.DefaultBufLength)
		}()

		select {
		case <-ch:
			// ctrl c pressed

			if err := dev.CancelAsync(); err != nil {
				log.Printf("CancelAsync failed - %s\n", err)
			} else {
				log.Printf("CancelAsync successful\n")
				log.Println("error", <-errch) // wait ReadAsync to return
			}
		case err := <-errch:
			if err != nil {
				log.Println("error", err)
			}
		}

	}
	log.Println("exiting ...")
}

func ppm_report(nsamples, interval uint64) int {

	real_rate := float64(nsamples) * 1e9 / float64(interval)
	ppm := 1e6 * (real_rate/float64(samp_rate) - 1)
	return int(ppm + 1)
}

var (
	startTime    time.Time
	totalSamples uint64
	nextInterval time.Duration
	skip         int = 20
)

func ppm_test(len int) {
	if skip > 0 {
		skip--
		return
	}
	// report ppm error every 10 sec
	const ppmIntervel = time.Second * 10

	if totalSamples == 0 {
		startTime = time.Now()
		totalSamples = uint64(len)
		nextInterval = ppmIntervel
		return
	}

	interval := time.Now().Sub(startTime)

	if interval < nextInterval {
		totalSamples += uint64(len)
		return
	}
	nextInterval += ppmIntervel

	fmt.Println("real sample rate", (1000000000*totalSamples)/uint64(interval), "ppm ", ppm_report(totalSamples, uint64(interval)))
	totalSamples += uint64(len)

}
