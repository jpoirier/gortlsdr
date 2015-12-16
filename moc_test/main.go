// Copyright (c) 2015-2016 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// +build ignore

package main

import (
	"log"

	rtl "./gortlsdr"
)

func GetDeviceCount() int {
	return rtl.GetDeviceCount()
}

func GetDeviceName(i int) string {
	return rtl.GetDeviceName(i)
}

func GetDeviceUsbStrings(i int) (m, p, s string, err error) {
	return rtl.GetDeviceUsbStrings(i)
}

func GetIndexBySerial(s string) (i int, err error) {
	return rtl.GetIndexBySerial(s)
}

func Open(i int) (*rtl.Context, error) {
	return rtl.Open(i)
}

func Close(d *rtl.Context) error {
	return d.Close()
}

func SetXtalFreq(d *rtl.Context, i int) {
	rtlFreqHz := 2500000
	tunerFreqHz := 5000000
	if err := d.SetXtalFreq(rtlFreqHz, tunerFreqHz); err != nil {
		log.Printf("--- FAILED, SetXtalFreq i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetXtalFreq i:%d\n", i)
	}
}

func GetXtalFreq(d *rtl.Context, i int) {
	if _, _, err := d.GetXtalFreq(); err != nil {
		log.Printf("--- FAILED, GetXtalFreq i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, GetXtalFreq i:%d\n", i)
	}
}

// func GetUsbStrings(d *rtl.Context, i int) {
// 	d.GetUsbStrings()
// }

// func WriteEeprom(d *rtl.Context, i int) {
// 	d.WriteEeprom(data, offset, leng)
// }

// func ReadEeprom(d *rtl.Context, i int) {
// 	d.ReadEeprom(data, offset, leng)
// }

func SetCenterFreq(d *rtl.Context, i int) {
	freqHz := 2500100
	if err := d.SetCenterFreq(freqHz); err != nil {
		log.Printf("--- FAILED, SetCenterFreq i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetCenterFreq i:%d\n", i)
	}
}

func GetCenterFreq(d *rtl.Context, i int) {
	if freqHz := d.GetCenterFreq(); freqHz == 0 {
		log.Printf("--- FAILED, GetCenterFreq i:%d - %d\n", i, freqHz)
	} else {
		log.Printf("--- PASSED, GetCenterFreq i:%d\n", i)
	}
}

func SetFreqCorrection(d *rtl.Context, i int) {
	ppm := 112
	if err := d.SetFreqCorrection(ppm); err != nil {
		log.Printf("--- FAILED, SetFreqCorrection i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetFreqCorrection i:%d\n", i)
	}
}

func GetFreqCorrection(d *rtl.Context, i int) {
	if ppm := d.GetFreqCorrection(); ppm != -1 {
		log.Printf("--- FAILED, GetFreqCorrection i:%d - %d\n", i, ppm)
	} else {
		log.Printf("--- PASSED, GetFreqCorrection i:%d\n", i)
	}
}

func GetTunerType(d *rtl.Context, i int) {
	tunerType := d.GetTunerType()
	switch tunerType {
	case "UNKNOWN", "RTLSDR_TUNER_UNKNOWN":
		log.Printf("--- FAILED, GetTunerType i:%d - %s\n", i, tunerType)
	default:
		log.Printf("--- PASSED, GetTunerType i:%d\n", i)
	}
}

func GetTunerGains(d *rtl.Context, i int) {
	if _, err := d.GetTunerGains(); err != nil {
		log.Printf("--- FAILED, GetTunerGains i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, GetTunerGains i:%d\n", i)
	}
}

func SetTunerGain(d *rtl.Context, i int) {
	gainTenthsDb := 110
	if err := d.SetTunerGain(gainTenthsDb); err != nil {
		log.Printf("--- FAILED, SetTunerGain i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetTunerGain i:%d\n", i)
	}
}

func SetTunerBw(d *rtl.Context, i int) {
	bwHz := 2300000
	if err := d.SetTunerBw(bwHz); err != nil {
		log.Printf("--- FAILED, SetTunerBw i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetTunerBw i:%d\n", i)
	}
}

// func GetTunerBw(d *rtl.Context, i int) {
// 	d.GetTunerBw
// }

func GetTunerGain(d *rtl.Context, i int) {
	if gainTenthsDb := d.GetTunerGain(); gainTenthsDb <= 0 {
		log.Printf("--- FAILED, GetTunerGain i:%d - %d\n", i, gainTenthsDb)
	} else {
		log.Printf("--- PASSED, GetTunerGain i:%d\n", i)
	}
}

func SetTunerIfGain(d *rtl.Context, i int) {
	stage := 6 // 1 - 6
	gainTenthsDb := -30
	if err := d.SetTunerIfGain(stage, gainTenthsDb); err != nil {
		log.Printf("--- FAILED, SetTunerIfGain i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetTunerIfGain i:%d\n", i)
	}
}

func SetTunerGainMode(d *rtl.Context, i int) {
	manualMode := true
	if err := d.SetTunerGainMode(manualMode); err != nil {
		log.Printf("--- FAILED, SetTunerGainMode i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetTunerGainMode i:%d\n", i)
	}
}

func SetSampleRate(d *rtl.Context, i int) {
	rateHz := 3500000
	if err := d.SetSampleRate(rateHz); err != nil {
		log.Printf("--- FAILED, SetSampleRate i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetSampleRate i:%d\n", i)
	}
}

func GetSampleRate(d *rtl.Context, i int) {
	if rateHz := d.GetSampleRate(); rateHz <= 0 {
		log.Printf("--- FAILED, GetSampleRate i:%d - %d\n", i, rateHz)
	} else {
		log.Printf("--- PASSED, GetSampleRate i:%d\n", i)
	}
}

func SetTestMode(d *rtl.Context, i int) {
	testMode := false
	if err := d.SetTestMode(testMode); err != nil {
		log.Printf("--- FAILED, SetTestMode i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetTestMode i:%d\n", i)
	}
}

func SetAgcMode(d *rtl.Context, i int) {
	AGCMode := true
	if err := d.SetAgcMode(AGCMode); err != nil {
		log.Printf("--- FAILED, SetAgcMode i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetAgcMode i:%d\n", i)
	}
}

func SetDirectSampling(d *rtl.Context, i int) {
	mode := rtl.SamplingIADC
	if err := d.SetDirectSampling(mode); err != nil {
		log.Printf("--- FAILED, SetDirectSampling i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetDirectSampling i:%d\n", i)
	}
}

func GetDirectSampling(d *rtl.Context, i int) {
	if _, err := d.GetDirectSampling(); err != nil {
		log.Printf("--- FAILED, GetDirectSampling i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, GetDirectSampling i:%d\n", i)
	}
}

func SetOffsetTuning(d *rtl.Context, i int) {
	enable := false
	if err := d.SetOffsetTuning(enable); err != nil {
		log.Printf("--- FAILED, SetOffsetTuning i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, SetOffsetTuning i:%d\n", i)
	}
}

func GetOffsetTuning(d *rtl.Context, i int) {
	if _, err := d.GetOffsetTuning(); err != nil {
		log.Printf("--- FAILED, GetOffsetTuning i:%d - %s\n", i, err)
	} else {
		log.Printf("--- PASSED, GetOffsetTuning i:%d\n", i)
	}
}

// func ResetBuffer(d *rtl.Context, i int) {
// 	d.ResetBuffer()
// }

// func ReadSync(d *rtl.Context, i int) {

// }

// func ReadAsync(d *rtl.Context, i int) {

// }

// func CancelAsync(d *rtl.Context, i int) {

// }

// func GetHwInfo(d *rtl.Context, i int) {

// }

// func SetHwInfo(d *rtl.Context, i int) {

// }

func main() {
	var cnt int

	//---------- Device Check ----------
	if cnt = GetDeviceCount(); cnt == 0 {
		log.Println("--- FAILED, GetDeviceCount no devices found, exiting...")
	} else {
		log.Println("--- PASSED, GetDeviceCount...")
	}

	for i := 0; i < cnt; i++ {
		if len(GetDeviceName(i)) == 0 {
			log.Printf("--- FAILED, GetDeviceName i:%d\n")
		} else {
			log.Printf("--- PASSED, GetDeviceName i:%d\n")
		}
	}

	for i := 3; i < 6; i++ {
		if len(GetDeviceName(i)) != 0 {
			log.Printf("--- FAILED, GetDeviceName i:%d\n")
		} else {
			log.Printf("--- PASSED, GetDeviceName i:%d\n")
		}
	}

	serials := make([]string, 3)
	for i := 0; i < cnt; i++ {
		if _, _, s, err := GetDeviceUsbStrings(i); err != nil {
			log.Printf("--- FAILED, GetDeviceUsbStrings i:%d, %s\n", i, err)
		} else {
			serials = append(serials, s)
			log.Printf("--- PASSED, GetDeviceUsbStrings i:%d\n")
		}
	}

	for i := 3; i < 6; i++ {
		if _, _, _, err := GetDeviceUsbStrings(i); err == nil {
			log.Printf("--- FAILED, GetDeviceUsbStrings i:%d, %s\n", i, err)
		} else {
			log.Printf("--- PASSED, GetDeviceUsbStrings i:%d\n")
		}
	}

	for i, s := range serials {
		if _, err := GetIndexBySerial(s); err != nil {
			log.Printf("--- FAILED, GetIndexBySerial i:%d - %s\n", i, err)
		} else {
			log.Printf("--- PASSED, GetIndexBySerial i:%d\n", i)
		}
	}

	for i, s := range []string{"One", "Two", "Three"} {
		if _, err := GetIndexBySerial(s); err == nil {
			log.Printf("--- FAILED, GetIndexBySerial i:%d - %s\n", i, err)
		} else {
			log.Printf("--- PASSED, GetIndexBySerial i:%d\n", i)
		}
	}

	for i := 5; i < 10; i++ {
		if d, err := rtl.Open(i); err == nil {
			log.Printf("--- FAILED, Open i:%d - %s\n", i, err)
			continue
		} else {
			d.Close()
			log.Printf("--- PASSED, Open i:%d\n", i)
		}
	}

	for i := 0; i < cnt; i++ {
		var err error
		var d *rtl.Context

		if d, err = rtl.Open(i); err != nil {
			log.Printf("--- FAILED, Open i:%d - %s\n", i, err)
			continue
		} else {
			log.Printf("--- PASSED, Open i:%d\n", i)
		}

		GetXtalFreq(d, i)
		SetXtalFreq(d, i)

		// GetUsbStrings(d, i)

		// ReadEeprom(d, i)
		// WriteEeprom(d, i)

		GetCenterFreq(d, i)
		SetCenterFreq(d, i)

		GetFreqCorrection(d, i)
		SetFreqCorrection(d, i)

		GetTunerType(d, i)

		GetTunerGains(d, i)
		SetTunerGain(d, i)

		// GetTunerBw(d, i)
		SetTunerBw(d, i)

		GetTunerGain(d, i)
		SetTunerIfGain(d, i)

		SetTunerGainMode(d, i)

		GetSampleRate(d, i)
		SetSampleRate(d, i)

		SetTestMode(d, i)
		SetAgcMode(d, i)

		GetDirectSampling(d, i)
		SetDirectSampling(d, i)

		GetOffsetTuning(d, i)
		SetOffsetTuning(d, i)

		// ResetBuffer(d, i)

		// GetHwInfo(d, i)
		// SetHwInfo(d, i)

		// ReadSync(d, i)
		// ReadAsync(d, i)

		// CancelAsync(d, i)

		if err = d.Close(); err != nil {
			log.Printf("--- FAILED, Close %s - %s...\n", err, i)
		} else {
			log.Println("--- PASSED, Close...")
		}
	}
}
