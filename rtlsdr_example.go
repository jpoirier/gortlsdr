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

	g := dev.GetTunerGains()
	for i, j := range g {
		log.Printf("Gain %d: %d\n", i, j)
	}

	log.Printf("Setting sample rate to %d\n", rtl.DEFAULT_SAMPLE_RATE)
	err = dev.SetSampleRate(rtl.DEFAULT_SAMPLE_RATE)
	if err != 0 {
		log.Fatal("SetSampleRate failed\n")
	}

	log.Printf("Closing...\n")
	dev.Close()
}
