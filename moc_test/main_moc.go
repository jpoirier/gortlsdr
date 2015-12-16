// Copyright (c) 2015-2016 Joseph D Poirier
// Distributable under the terms of The New BSD License
// that can be found in the LICENSE file.

// +build ignore

package main

import (
	"errors"
	"log"

	rtl "./gortlsdr_moc"
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
	return rtl.GetIndexBySerial(i)
}

func Open(i int) (*rtl.Context, error){
	return rtl.Open(i)
}

func Close(d *rtl.Context) error {
	return d.Close()
}

func SetXtalFreq(d *rtl.Context) {

}

func GetXtalFreq(d *rtl.Context) {

}

func GetUsbStrings(d *rtl.Context) {

}

func WriteEeprom(d *rtl.Context) {

}

func ReadEeprom(d *rtl.Context) {

}

func SetCenterFreq(d *rtl.Context) {

}

func GetCenterFreq(d *rtl.Context) {

}

func SetFreqCorrection(d *rtl.Context) {

}

func GetFreqCorrection(d *rtl.Context) {

}

func GetTunerType(d *rtl.Context) {

}

func GetTunerGains(d *rtl.Context) {

}

func SetTunerGains(d *rtl.Context) {

}

func SetTunerBw(d *rtl.Context) {

}

func GetTunerBw(d *rtl.Context) {

}

func GetTunerGain(d *rtl.Context) {

}

func SetTunerIfGain(d *rtl.Context) {

}

func SetTunerGainMode(d *rtl.Context) {

}

func SetSampleRate(d *rtl.Context) {

}

func GetSampleRate(d *rtl.Context) {

}

func SetTestMode(d *rtl.Context) {

}

func SetAgcMode(d *rtl.Context) {

}

func SetDirectSampling(d *rtl.Context) {

}

func GetDirectSampling(d *rtl.Context) {

}

func SetOffsetTuning(d *rtl.Context) {

}

func GetOffsetTuning(d *rtl.Context) {

}

func ResetBuffer(d *rtl.Context) {

}

func ReadSync(d *rtl.Context) {

}

func ReadAsync(d *rtl.Context) {

}

func CancelAsync(d *rtl.Context) {

}

func GetHwInfo(d *rtl.Context) {
}

func SetHwInfo(d *rtl.Context) {
}

func main() {
	//---------- Device Check ----------
	if c := GetDeviceCount(); c == 0 {
		log.Fatal("No devices found, exiting.\n")
	}
	for i := 0; i < c; i++ {
		if m, p, s, err := GetDeviceUsbStrings(i); if err == nil {
			err = errors.New("")
			log.Printf("GetDeviceUsbStrings %s - %s %s %s\n", err, m, p, s)
		}
	}

}
