/*
 * @filename	powermeter.go
 * @author		kontornl
 * @date		19/07/2024
 * @desc		Constants and methods of Modbus-RTU electric meter
 * @comment		--
 */

package powermeter

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
	// voltage, in Vrms
	ID_VOLTAGE uint8 = iota
	// phase A voltage, in Vrms
	ID_VOLTAGE_PHASEA
	// phase B voltage, in Vrms
	ID_VOLTAGE_PHASEB
	// phase C voltage, in Vrms
	ID_VOLTAGE_PHASEC

	// voltage, in Arms
	ID_CURRENT
	// phase A voltage, in Arms
	ID_CURRENT_PHASEA
	// phase B voltage, in Arms
	ID_CURRENT_PHASEB
	// phase C voltage, in Arms
	ID_CURRENT_PHASEC

	// active power, in W
	ID_POWER_ACTIVE
	// phase A active power, in W
	ID_POWER_ACTIVE_PHASEA
	// phase B active power, in W
	ID_POWER_ACTIVE_PHASEB
	// phase C active power, in W
	ID_POWER_ACTIVE_PHASEC

	// passive power, in var
	ID_POWER_PASSIVE
	// phase A passive power, in var
	ID_POWER_PASSIVE_PHASEA
	// phase B passive power, in var
	ID_POWER_PASSIVE_PHASEB
	// phase C passive power, in var
	ID_POWER_PASSIVE_PHASEC

	// apparent power, in VA
	ID_POWER_APPARENT
	// phase A apparent power, in VA
	ID_POWER_APPARENT_PHASEA
	// phase B apparent power, in VA
	ID_POWER_APPARENT_PHASEB
	// phase C apparent power, in VA
	ID_POWER_APPARENT_PHASEC

	// power factor, in 1.0 (ranged 0 - 1)
	ID_POWER_FACTOR
	// phase A power factor, in 1.0 (ranged 0 - 1)
	ID_POWER_FACTOR_PHASEA
	// phase B power factor, in 1.0 (ranged 0 - 1)
	ID_POWER_FACTOR_PHASEB
	// phase C power factor, in 1.0 (ranged 0 - 1)
	ID_POWER_FACTOR_PHASEC

	// power line frequency, in Hz
	ID_FREQ

	// indicating value of current active energy of all rates, in kWh
	ID_ENERGY_ACTIVE_CURR_ALL
	// indicating value of current positive active energy of all rates, in kWh
	ID_ENERGY_ACTIVE_POSI_CURR_ALL
	// indicating value of current negative active energy of all rates, in kWh
	ID_ENERGY_ACTIVE_NEGA_CURR_ALL
	// indicating value of current passive energy of all rates, in kWh
	ID_ENERGY_PASSIVE_CURR_ALL
	// indicating value of current positive passive energy of all rates, in kWh
	ID_ENERGY_PASSIVE_POSI_CURR_ALL
	// indicating value of current negative passive energy of all rates, in kWh
	ID_ENERGY_PASSIVE_NEGA_CURR_ALL

	// Modbus-RTU slave address, 1 byte (ranged 1 - 247)
	ID_SLAVE_ADDR
	// date and time
	ID_DATETIME

	// (reserved) ID amount counter, must be at the end
	ID_DATA_ITEM_AMOUNT__
)

const (
	POWERSWITCH_TURN_1 = iota
	POWERSWITCH_TURN_2
	POWERSWITCH_TURN_3
	POWERSWITCH_TURN_4
	POWERSWITCH_TURN_5
	POWERSWITCH_TURN_6
	POWERSWITCH_TURN_7
	POWERSWITCH_TURN_8
)

// meter model definitions
const (
	METER_MODEL_DDS4921 uint8 = iota
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
func (pm *PowerMeter) Init(gw *gateway.MBRTGateway, meterModel uint8, slaveAddr uint8) (err error) {
	if slaveAddr > 60 {
		err = errors.New("invalid slave address which exceeds 60")
		return
	}
	if meterModel == METER_MODEL_DDS4921 {
		pm.regMeta = regMetaDDS4921
		pm.SwitchMeta = switchMetaDDS4921
	} else {
		// invalid meter model
		err = errors.New("invalid meter type")
		return
	}
	pm.gateway = gw
	pm.slaveAddr = slaveAddr
	return
}

/*
get values such as voltage, power and energy

# Params

id uint8: item id, specifies which value should be fetched, using macro ID_*

# Returns

ret float64: value in float64, the unit might be one of the following: Vrms, Arms, W, var, VA, Hz, kWh, kvarh

err error: error
*/
func (pm *PowerMeter) GetVal(id uint8) (ret float64, err error) {
	time.Sleep(5 * time.Millisecond)
	if pm.gateway.GetClient() == nil {
		pm.gateway.Reinit()
	}
	pm.gateway.GetClient().SetUnitId(pm.slaveAddr)
	var regval []uint16
	ret = 0.0
	if pm.regMeta[id].length == 0 {
		err = errors.New("undefined register metadata")
		return
	}
	if !pm.regMeta[id].readable {
		err = errors.New("unreadable register")
		return
	}
	for retry := 3; ; retry-- {
		regval, err = pm.gateway.GetClient().ReadRegisters(
			pm.regMeta[id].regAddr,
			pm.regMeta[id].length,
			modbus.HOLDING_REGISTER,
		)
		if err != nil {
			if retry > 0 {
				err = pm.gateway.Reconnect()
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
	ret *= float64(pm.regMeta[id].override)
	if pm.regMeta[id].hasSymbol && (regval[0]/32768 == 1) {
		ret *= -1
	}
	return
}

/*
fetch power switch status

# Params

turn uint8: which switch should be operated, using macro POWERSWITCH_TURN_*

# Returns

stat bool: switch status, true is closed (turned on), false is tripped (turned off)

err error: error
*/
func (pm *PowerMeter) GetSwitchStatus(turn uint8) (stat bool, err error) {
	time.Sleep(5 * time.Millisecond)
	if pm.gateway.GetClient() == nil {
		pm.gateway.Reinit()
	}
	pm.gateway.GetClient().SetUnitId(pm.slaveAddr)
	var regval uint16
	stat = false
	// (23/07/2024 kontornl) the register may just a coil, not a holding register
	for retry := 3; ; retry-- {
		regval, err = pm.gateway.GetClient().ReadRegister(pm.SwitchMeta[turn].statusAddr, modbus.HOLDING_REGISTER)
		if err != nil {
			if retry > 0 {
				err = pm.gateway.Reconnect()
			}
			if err != nil {
				return
			}
		} else {
			break
		}
	}
	if regval == pm.SwitchMeta[turn].statusTripVal {
		stat = false
	} else if regval == pm.SwitchMeta[turn].statusCloseVal {
		stat = true
	} else {
		err = fmt.Errorf("bad register value 0x%04x", regval)
	}
	return
}

/*
power switch trip (turn off) command

# Params

turn uint8: which switch should be operated, give operand using macro such as POWERSWITCH_TURN_1

# Returns

err error: error
*/
func (pm *PowerMeter) Trip(turn uint8) (err error) {
	time.Sleep(5 * time.Millisecond)
	if pm.gateway.GetClient() == nil {
		pm.gateway.Reinit()
	}
	pm.gateway.GetClient().SetUnitId(pm.slaveAddr)
	for retry := 3; ; retry-- {
		err = pm.gateway.GetClient().WriteRegisters(pm.SwitchMeta[turn].ctlAddr, []uint16{pm.SwitchMeta[turn].ctlTripCmd})
		if err != nil {
			if retry > 0 {
				err = pm.gateway.Reconnect()
			}
			if err != nil {
				return
			}
		} else {
			time.Sleep(200 * time.Millisecond)
			break
		}
	}
	var stat bool
	stat, err = pm.GetSwitchStatus(turn)
	if err != nil {
		return
	}
	if stat != false {
		err = errors.New("power switch status mismatch")
	}
	return
}

/*
power switch close (turn on) command

# Params

turn uint8: which switch should be operated, give operand using macro such as POWERSWITCH_TURN_1

# Returns

err error: error
*/
func (pm *PowerMeter) Close(turn uint8) (err error) {
	time.Sleep(5 * time.Millisecond)
	if pm.gateway.GetClient() == nil {
		pm.gateway.Reinit()
	}
	pm.gateway.GetClient().SetUnitId(pm.slaveAddr)
	for retry := 3; ; retry-- {
		err = pm.gateway.GetClient().WriteRegisters(pm.SwitchMeta[turn].ctlAddr, []uint16{pm.SwitchMeta[turn].ctlCloseCmd})
		if err != nil {
			if retry > 0 {
				err = pm.gateway.Reconnect()
			}
			if err != nil {
				return
			}
		} else {
			time.Sleep(200 * time.Millisecond)
			break
		}
	}
	var stat bool
	stat, err = pm.GetSwitchStatus(turn)
	if err != nil {
		return
	}
	if stat != true {
		err = errors.New("power switch status mismatch")
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

// metadata of power switch
type SwitchMeta struct {
	// switch controlling register address
	ctlAddr uint16
	// switch trip command to write to register
	ctlTripCmd uint16
	// switch close command to write to register
	ctlCloseCmd uint16
	// switch status register address
	statusAddr uint16
	// value indicates that switch is tripped
	statusTripVal uint16
	// value indicates that switch is closed
	statusCloseVal uint16
}

type PowerMeter struct {
	gateway    *gateway.MBRTGateway
	slaveAddr  uint8
	regMeta    []RegMeta
	SwitchMeta []SwitchMeta
}

type IPowerMeter interface {
	Init(gw *gateway.MBRTGateway, meterModel uint8, slaveAddr uint8) (err error)
	GetVal(id uint8) (ret float64, err error)
	GetSwitchStatus(turn uint8) (stat bool, err error)
	Trip(turn uint8) (err error)
	Close(turn uint8) (err error)
}
