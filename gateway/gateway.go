package gateway

import (
	"time"

	"github.com/kontornl/modbus"
)

/*
initialize gateway instance and establish TCP connection

multiple meters may connected to one gateway and identified by their slave address, which is not specified here

# Params

netAddr string: the gateway to connect to, string format is rtuovertcp://<ip>:<port>

baudRate uint: serial baud rate which is already set to gateway

timeout time.Duration: operation time-out

# Returns

err error: error
*/
func (gw *MBRTGateway) Init(netAddr string, baudRate uint, timeout time.Duration) (err error) {
	var cli *modbus.ModbusClient
	if gw.cli != nil {
		// (23/07/2024 kontornl) may cause memory leak, need inspection
		err := gw.cli.Close()
		if err != nil {
			return err
		}
	}
	cli, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:     netAddr,
		Speed:   baudRate,
		Timeout: timeout,
	})
	if err != nil {
		return
	}
	err = cli.Open()
	if err != nil {
		return err
	}
	gw.cli = cli
	return
}

func (gw *MBRTGateway) Reconnect() (err error) {
	gw.cli.Close()
	err = gw.cli.Open()
	return
}

func (gw *MBRTGateway) GetClient() (cli *modbus.ModbusClient) {
	cli = gw.cli
	return
}

type MBRTGateway struct {
	cli *modbus.ModbusClient
}

type IMBRTGateway interface {
	Init(netAddr string, baudRate uint, timeout time.Duration) (err error)
	Reconnect() (err error)
	GetClient() (cli *modbus.ModbusClient)
}
