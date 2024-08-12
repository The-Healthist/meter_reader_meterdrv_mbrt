/*
 * @filename	watermeter.go
 * @author		kontornl
 * @date		24/07/2024
 * @desc		Constants and methods of Modbus-RTU water meter
 * @comment		--
 */

package watermeter

import (
	"errors"
	"fmt"
	"time"

	"github.com/The-Healthist/meter_reader_meterdrv_mbrt/gateway"

	"github.com/kontornl/modbus"
)

// data item identifiers in serial form
// (23/07/2024 kontornl) 常量封装到结构体中，使用x.x形式访问
const (
	// indicating number of water volume, in m^3
	ID_VOLUME = iota
	// (reserved) ID amount counter, must be at the end
	ID_DATA_ITEM_AMOUNT__
)

const (
	VALVE_TURN_1 uint8 = iota
	VALVE_TURN_2
	VALVE_TURN_3
	VALVE_TURN_4
)

// meter model identifiers
const (
	METER_MODEL_HYLSY uint8 = iota
)

// Modbus-RTU register type identifiers
const (
	REGTYPE_COIL = iota
	REGTYPE_INPUT
	REGTYPE_HOLDING
)

/*
initialize power meter instance

# Params

gw *gateway.MBRTGateway: the gateway instance that meter actually connect to, and needs to be initialized in advance

meterModel uint8: meter model id, using macro METER_MODEL_*

slaveAddr uint8: Modbus-RTU address of the meter

# Returns

err error: error
*/
func (wm *WaterMeter) Init(gw *gateway.MBRTGateway, meterModel uint8, slaveAddr uint8) (err error) {
	if slaveAddr > 60 {
		err = errors.New("invalid slave address which exceeds 60")
		return
	}
	if meterModel == METER_MODEL_HYLSY {
		wm.regMeta = regMetaHYLSY
		wm.valveMeta = valveMetaHYLSY
	}
	wm.gateway = gw
	wm.slaveAddr = slaveAddr
	return
}

/*
get values such as water volume

# Params

id uint8: item id, specifies which value should be fetched, using macro ID_*

# Returns

ret float64: value in float64, the unit might be one of the following: m^3

err error: error
*/
func (wm *WaterMeter) GetVal(id uint8) (ret float64, err error) {
	var regval []uint16
	ret = 0.0
	wm.gateway.GetClient().SetUnitId(wm.slaveAddr)
	if wm.regMeta[id].length == 0 {
		err = errors.New("undefined register metadata")
		return
	}
	if !wm.regMeta[id].readable {
		err = errors.New("unreadable register")
		return
	}
	for retry := 3; ; retry-- {
		regval, err = wm.gateway.GetClient().ReadRegisters(
			wm.regMeta[id].regAddr,
			wm.regMeta[id].length,
			modbus.HOLDING_REGISTER,
		)
		time.Sleep(50 * time.Millisecond)
		if err != nil {
			if retry > 0 {
				err = wm.gateway.Reconnect()
			}
			if err != nil {
				return
			}
		} else {
			break
		}
	}
	for i := 0; i < len(regval); i++ {
		ret *= 65536
		ret += float64(regval[i])
	}
	ret *= float64(wm.regMeta[id].override)
	if wm.regMeta[id].hasSymbol && (regval[0]/32768 == 1) {
		ret *= -1
	}
	return
}

/*
fetch valve status

# Params

turn uint8: which switch should be operated, using macro POWERSWITCH_TURN_*

# Returns

stat bool: switch status, true is opened (turned on), false is closed (turned off)

err error: error
*/
func (wm *WaterMeter) GetValve(turn uint8) (stat bool, err error) {
	var regval uint16
	stat = false
	wm.gateway.GetClient().SetUnitId(wm.slaveAddr)
	// (23/07/2024 kontornl) the register may just a coil, not a holding register
	for retry := 3; ; retry-- {
		if wm.valveMeta[turn].statusRegType == REGTYPE_COIL {
			stat, err = wm.gateway.GetClient().ReadCoil(wm.valveMeta[turn].statusAddr)
		} else if wm.valveMeta[turn].statusRegType == REGTYPE_HOLDING {
			regval, err = wm.gateway.GetClient().ReadRegister(wm.valveMeta[turn].statusAddr, modbus.HOLDING_REGISTER)
			if regval == wm.valveMeta[turn].statusCloseVal {
				stat = false
			} else if regval == wm.valveMeta[turn].statusOpenVal {
				stat = true
			} else {
				err = fmt.Errorf("bad register value 0x%04x", regval)
			}
		} else {
			err = errors.New("invalid register type")
			return
		}
		time.Sleep(50 * time.Millisecond)
		if err != nil {
			if retry > 0 {
				err = wm.gateway.Reconnect()
			}
			if err != nil {
				return
			}
		} else {
			break
		}
	}
	return
}

/*
valve action (open or close) command

# Params

turn uint8: which switch should be operated, give operand using macro such as POWERSWITCH_TURN_1

stat bool: switch status, true is opened (turned on), false is closed (turned off)

# Returns

err error: error
*/
func (wm *WaterMeter) SetValve(turn uint8, stat bool) (err error) {
	wm.gateway.GetClient().SetUnitId(wm.slaveAddr)
	for retry := 30; ; retry-- {
		if wm.valveMeta[turn].statusRegType == REGTYPE_COIL {
			err = wm.gateway.GetClient().WriteCoil(
				wm.valveMeta[turn].ctlAddr,
				stat,
			)
		} else if wm.valveMeta[turn].statusRegType == REGTYPE_HOLDING {
			var cmd uint16
			if stat {
				cmd = wm.valveMeta[turn].ctlOpenCmd
			} else {
				cmd = wm.valveMeta[turn].ctlCloseCmd
			}
			err = wm.gateway.GetClient().WriteRegisters(
				wm.valveMeta[turn].ctlAddr,
				[]uint16{cmd},
			)
		} else {
			err = errors.New("invalid register type")
			return
		}
		time.Sleep(50 * time.Millisecond)
		if err != nil {
			if retry > 0 {
				err = wm.gateway.Reconnect()
			}
			if err != nil {
				return
			}
		} else {
			break
		}
	}
	// time.Sleep(6 * time.Second)
	var newstat bool
	newstat, err = wm.GetValve(turn)
	if err != nil {
		return
	}
	if stat != newstat {
		err = errors.New("water switch status mismatch")
	}
	return
}

// metadata of registers in meter, including reg addr, length, read/writability and so on
type RegMeta struct {
	// register address
	regAddr uint16
	// number of successive registers used to hold one value
	length uint16
	// if register readable
	readable bool
	// if register writable
	writable bool
	// if the value can be no less than 0
	hasSymbol bool
	// a value multiplied onto the original value from the register
	override float32
}

// metadata of valve
type ValveMeta struct {
	// valve controlling register address
	ctlAddr uint16
	// valve controlling register type (coil or holding)
	ctlRegType uint8
	// valve close command to write to register
	ctlCloseCmd uint16
	// valve open command to write to register
	ctlOpenCmd uint16
	// valve status register address
	statusAddr uint16
	// valve status register type (coil or holding)
	statusRegType uint8
	// value indicates that valve is closed
	statusCloseVal uint16
	// value indicates that valve is opened
	statusOpenVal uint16
}

type WaterMeter struct {
	gateway   *gateway.MBRTGateway
	slaveAddr uint8
	regMeta   []RegMeta
	valveMeta []ValveMeta
}

type IWaterMeter interface {
	Init(gw *gateway.MBRTGateway, meterModel uint8, slaveAddr uint8) (err error)
	GetVal(id uint8) (ret float64, err error)
	GetValve(turn uint8) (stat bool, err error)
	SetValve(turn uint8, stat bool) (err error)
}
