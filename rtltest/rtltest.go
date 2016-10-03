package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	// "runtime"

	rtl "github.com/jpoirier/gortlsdr"
)

// report ppm error every 10 sec
const ppmInterval = time.Second * 10

func ConfigTest(dev *rtl.Context, samplerate, ppmErr int) error {

	//---------- Device Strings ----------
	m, p, s, err := dev.GetUsbStrings()
	if err != nil {
		log.Printf("\tGetUsbStrings Failed - error: %s\n", err)
		return err
	}
	fmt.Printf("\tGetUsbStrings - %s %s %s\n", m, p, s)
	fmt.Printf("\tGetTunerType: %s\n", dev.GetTunerType())

	//---------- Device Info ----------
	info, err := dev.GetHwInfo()
	if err != nil {
		log.Printf("\tGetHwInfo Failed - error: %s\n", err)
		return err
	}
	fmt.Printf("\tVendor ID      : 0x%X\n", info.VendorID)
	fmt.Printf("\tProduct ID     : 0x%X\n", info.ProductID)
	fmt.Println("\tManufacturer  : ", info.Manufact)
	fmt.Println("\tProduct       : ", info.Product)
	fmt.Println("\tSerial        : ", info.Serial)
	fmt.Println("\tHave Serial   : ", info.HaveSerial)
	fmt.Println("\tEnable IR     : ", info.EnableIR)
	fmt.Println("\tRemote Wakeup : ", info.RemoteWakeup)

	if ppmErr != 0 {
		err = dev.SetFreqCorrection(ppmErr)
		if err != nil {
			log.Printf("\tSetFreqCorrection %d Failed, error: %s\n", ppmErr, err)
			return err
		}
		fmt.Printf("\tSetFreqCorrection %d Successful\n", ppmErr)
	}

	//---------- Get/Set Sample Rate ----------
	err = dev.SetSampleRate(samplerate)
	if err != nil {
		log.Printf("\tSetSampleRate Failed - error: %s\n", err)
		return err
	}

	fmt.Printf("SampleRate: %d\n", dev.GetSampleRate())

	if err = dev.SetTestMode(true); err != nil {
		log.Printf("\tSetTestMode 'On' Failed - error: %s\n", err)
		return err
	}
	fmt.Printf("\tSetTestMode 'On' Successful\n")

	if err = dev.ResetBuffer(); err != nil {
		log.Printf("\tResetBuffer Failed - error: %s\n", err)
		return err
	}
	// fmt.Printf("\tResetBuffer Successful\n")

	return nil
}

func rtlsdrCb(dev *rtl.Context, buf []byte, u interface{}) {
	c := u.(*config)
	if !c.first {
		c.first = true
		//c.bcnt = buf[0]
		c.bcnt = buf[len(buf)-1] + 1
		return //skip firs buffer
	}
	lost := 0

	for _, v := range buf {
		if c.bcnt != v {
			//fmt.Println(f, i, v, bcnt)
			if v > c.bcnt {
				lost += int(v - c.bcnt)
			} else {
				lost += 256 + int(c.bcnt-v)
			}

			c.clost++
			c.glost += lost
			c.bcnt = v
		}
		c.bcnt++
	}
	if lost > 0 {
		c.out <- fmt.Sprintf("%v, lost at least %d bytes", c.device, lost)
	}
	ppm_test(len(buf)/2, c)
}

type config struct {
	dev        *rtl.Context
	device     int
	samplerate int
	ppmErr     int
	syncTest   bool

	// test vars
	first        bool
	bcnt         uint8
	glost, clost int
	samp_rate    int

	startTime    time.Time
	totalSamples uint64
	nextInterval time.Duration
	ppmSkip      int

	out chan string
}

func setup(c *config) error {
	fmt.Printf("%d ===== Device name: %s =====\n", c.device, rtl.GetDeviceName(c.device))
	var err error
	c.dev, err = rtl.Open(c.device)
	if err != nil {
		log.Println(c.device, "rtl.Open return error", err)
		return err
	}

	err = ConfigTest(c.dev, c.samplerate, c.ppmErr)
	if err != nil {
		log.Println("Config failed:", err.Error())
		c.dev.Close()
		return err
	}
	return nil
}

func testA(quit chan struct{}, c *config, wg *sync.WaitGroup) {

	defer wg.Done()

	if c.syncTest {
		var buffer = make([]uint8, rtl.DefaultBufLength)
	FOR:
		for {

			n, err := c.dev.ReadSync2(buffer)
			if err != nil {
				log.Println("ReadSync2 error", err.Error())
				break FOR
			}
			if n != rtl.DefaultBufLength {
				log.Println("short sync read")
				break FOR
			}
			rtlsdrCb(c.dev, buffer[:n], c)

			select {
			case <-quit:
				break FOR
			default:
			}
		}
	} else {

		errch := make(chan error)
		// if in goroutine that call ReadAsync is call to defer dev.Close(),
		// in case of panic in callback, program deadlock.
		wg.Add(1)
		go func() {
			defer wg.Done()
			// rtl-sdr error: Failed to submit transfer 31!
			// http://www.rtl-sdr.com/forum/viewtopic.php?f=7&t=142
			errch <- c.dev.ReadAsync2(rtlsdrCb, c, rtl.DefaultAsyncBufNumber/4, rtl.DefaultBufLength)
		}()

		select {
		case <-quit:
			if err := c.dev.CancelAsync(); err != nil {
				log.Printf("CancelAsync failed - %s\n", err)
			} else {
				fmt.Printf("CancelAsync successful\n")
				err = <-errch // wait ReadAsync to return
				if err != nil {
					log.Println("ReadAsync2 error", err)
				}
			}
		case err := <-errch:
			if err != nil {
				log.Println("error", err)
			}
		}

	}
	fmt.Println(c.device, "exiting ...")
}

func ppm_report(nsamples, interval uint64, c *config) int {

	real_rate := float64(nsamples) * 1e9 / float64(interval)
	ppm := 1e6 * (real_rate/float64(c.samplerate) - 1)
	return int(ppm + 1)
}

func ppm_test(len int, c *config) {
	if c.ppmSkip > 0 {
		c.ppmSkip--
		return
	}

	if c.totalSamples == 0 {
		c.startTime = time.Now()
		c.totalSamples = uint64(len)
		c.nextInterval = ppmInterval
		return
	}

	interval := time.Now().Sub(c.startTime)

	if interval < c.nextInterval {
		c.totalSamples += uint64(len)
		return
	}
	c.nextInterval += ppmInterval

	c.out <- fmt.Sprint(c.device, " real sample rate ", (1000000000*c.totalSamples)/uint64(interval), " ppm ", ppm_report(c.totalSamples, uint64(interval), c))
	c.totalSamples += uint64(len)

}

func output(o chan string) {
	for s := range o {
		fmt.Println(s)
	}
}

func main() {
	var device = flag.Int("d", 0, "device idx to use, if negative -n use first n devices")
	var samplerate = flag.Int("s", 2048000,
		"Sample rate Hz. The RTL2832U can sample from two ranges: 225001..300000 and 900001..3200000")
	var ppmErr = flag.Int("p", 0, "PPM Error")
	var syncTest = flag.Bool("S", false, "use sync calls")
	flag.Parse()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)

	//---------- Device Check ----------
	if c := rtl.GetDeviceCount(); c == 0 {
		log.Fatal("No devices found, exiting.\n")
	} else {
		fmt.Println("Devices found:")
		for i := 0; i < c; i++ {
			m, p, s, err := rtl.GetDeviceUsbStrings(i)
			if err != nil {
				log.Println(i, "GetDeviceUsbStrings return error", err)
				return
			} else {
				fmt.Printf("GetDeviceUsbStrings: %s %s %s\n", m, p, s)
			}
		}
	}

	// print all info from calbacks in separate goroutine
	str := make(chan string, 100)
	go output(str)

	var cfgs []*config

	if *device < 0 {
		fmt.Println("testing first", -*device, "devices")
		for i := 0; i < -*device; i++ {
			tc := &config{
				device:     i,
				samplerate: *samplerate,
				ppmErr:     *ppmErr,
				syncTest:   *syncTest,
				ppmSkip:    40,
				out:        str,
			}
			cfgs = append(cfgs, tc)
			if err := setup(tc); err != nil {
				log.Printf("setup device %d return error: %s", i, err)
				return

			}
		}

	} else {
		fmt.Println("testing device", *device)
		tc := &config{
			device:     *device,
			samplerate: *samplerate,
			ppmErr:     *ppmErr,
			syncTest:   *syncTest,
			ppmSkip:    40,
			out:        str,
		}
		cfgs = append(cfgs, tc)
		if err := setup(tc); err != nil {
			log.Printf("setup device %d return error: %s", *device, err)
			return

		}
	}
	fmt.Println("\nThis program compute ppm error(relative to -p parameter) every", ppmInterval)
	fmt.Println("If lost samples are reported, computed ppm is not correct,")
	fmt.Println(" and sample rate must be decreased, until no lost samples are reported\n")

	wg := &sync.WaitGroup{}
	quit := make(chan struct{})
	for i := range cfgs {
		wg.Add(1)
		go testA(quit, cfgs[i], wg)
	}

	select {
	case <-ch:
		// ctrl c pressed
	}
	close(quit)
	wg.Wait()

	for i := range cfgs {
		fmt.Printf("device %d, total lost bytes %d in %d buffers\n", i, cfgs[i].glost, cfgs[i].clost)

	}

}

/*
rtl_test
Found 2 device(s):
  0:  Realtek, RTL2838UHIDIR, SN: 00000001
  1:  Realtek, RTL2838UHIDIR, SN: 00000002

Using device 0: Generic RTL2832U OEM
Found Rafael Micro R820T tuner
Supported gain values (29): 0.0 0.9 1.4 2.7 3.7 7.7 8.7 12.5 14.4 15.7 16.6 19.7 20.7 22.9 25.4 28.0 29.7 32.8 33.8 36.4 37.2 38.6 40.2 42.1 43.4 43.9 44.5 48.0 49.6
[R82XX] PLL not locked!
Sampling at 2048000 S/s.

Info: This tool will continuously read from the device, and report if
samples get lost. If you observe no further output, everything is fine.

Reading samples in async mode...
^CSignal caught, exiting!

User cancel, exiting...
Samples per million lost (minimum): 0

*/
