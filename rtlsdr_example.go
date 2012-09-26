// +build ignore

package main

import (
	rtl "github.com/jpoirier/gortlsdr"
	"log"
)

func main() {
	indx := -1
	c := rtl.GetDeviceCount()
	if c == 0 {
		log.Fatal("No devices found.\n")
	}

	for i := 0; i < c; i++ {
		m, p, s, err := rtl.GetDeviceUsbStrings(i)
		if err != 0 {
			indx++
			log.Printf("err: %d, m: %s, p: %s, s: %s\n", err, m, p, s)
		}
	}

	if indx != -1 {
		log.Fatal("No devices found.\n")
	}

	log.Printf("Using device indx %d\n", 0)
	dev, err := rtl.Open(0)
	if err != 0 {
		log.Fatal("Failed to open the device\n")
	}
	defer dev.Close()

	g := dev.GetTunerGains()
	for i, j := range g {
		log.Printf("Gain %d: %d\n", i, j)
	}

	log.Printf("Setting sample rate to %d\n", rtl.DEFAULT_SAMPLE_RATE)
	err = dev.SetSampleRate(rtl.DEFAULT_SAMPLE_RATE)
	if err != 0 {
		log.Fatal("SetSampleRate failed, exiting\n")
	}

	err = dev.SetTestMode(1)
	if err == -1 {
		log.Fatal("Setting test mode failed, exiting\n")
	}

	err = dev.ResetBuffer()
	if err == -1 {
		log.Fatal("Buffer reset failed, exiting\n")
	}

	var buffer []byte = make([]uint8, rtl.DEFAULT_BUF_LENGTH)
	n_read, err := dev.ReadSync(buffer, rtl.DEFAULT_BUF_LENGTH)
	if err == -1 {
		log.Fatal("ReadSync failed, exiting\n")
	}
	if n_read < rtl.DEFAULT_BUF_LENGTH {
		log.Fatal("ReadSync short read, samples lost, exiting\n")
	}
	log.Println("ReadSync successful")
	// log.Println(buffer)

	log.Printf("Closing...\n")
}
