// +build ignore

package main

import (
	"log"
	"rtlsdr"
)

func main() {
	indx := -1
	c := rtlsdr.GetDeviceCount()

	for i := 0; i < int(c); i++ {
		m, p, s, err := rtlsdr.GetDeviceUsbStrings(uint32(i))
		if err != 0 {
			indx++
			log.Printf("err: %d, m: %s, p: %s, s: %s\n", err, m, p, s)
		}
	}

	if indx != -1 {
		log.Fatal("No devices found.\n")
	}

	log.Printf("Using device indx %d\n", 0)
	rtl, err := rtlsdr.Open(uint32(0))
	if err != 0 {
		log.Fatal("Failed to open the device\n")
	}

	g := rtl.GetTunerGains()
	for i := 0; i < len(g); i++ {
		log.Printf("Gain %d: %d\n", i, g[i])
	}

	log.Printf("Setting sample rate to %d\n", rtlsdr.DEFAULT_SAMPLE_RATE)
	err = rtl.SetSampleRate(rtlsdr.DEFAULT_SAMPLE_RATE)
	if err != 0 {
		log.Fatal("SetSampleRate failed\n")
	}

	log.Printf("Closing...\n")
	rtl.Close()
}
