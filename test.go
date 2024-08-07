package main

import (
	"fmt"
	"time"

	"github.com/The-Healthist/meter_reader_meterdrv_mbrt/gateway"
	"github.com/The-Healthist/meter_reader_meterdrv_mbrt/powermeter"
	"github.com/The-Healthist/meter_reader_meterdrv_mbrt/watermeter"
)

func main() {
	gw := new(gateway.MBRTGateway)
	pm := new(powermeter.PowerMeter)
	wm := new(watermeter.WaterMeter)
	err := gw.Init("rtuovertcp://192.168.1.12:8802", 9600, 5*time.Second)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = pm.Init(gw, powermeter.METER_MODEL_DDS4921, 0x02)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = wm.Init(gw, watermeter.METER_MODEL_HYLSY, 0x15)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	var ret float64
	ret, err = pm.GetVal(powermeter.ID_POWER_FACTOR)
	if err == nil {
		fmt.Printf("value: %3.03f\n", ret)
	} else {
		fmt.Printf("error: %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)
	ret, err = wm.GetVal(watermeter.ID_VOLUME)
	if err == nil {
		fmt.Printf("value: %3.03f\n", ret)
	} else {
		fmt.Printf("error: %v\n", err)
	}

	time.Sleep(3 * time.Second)

	err = pm.Trip(powermeter.POWERSWITCH_TURN_1)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)
	err = wm.SetValve(watermeter.VALVE_TURN_1, false)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	time.Sleep(3 * time.Second)

	err = pm.Close(powermeter.POWERSWITCH_TURN_1)
	if err != nil {
		// 	fmt.Printf("value: %3.02f", val)
		// } else {
		fmt.Printf("error: %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)
	err = wm.SetValve(watermeter.VALVE_TURN_1, true)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
