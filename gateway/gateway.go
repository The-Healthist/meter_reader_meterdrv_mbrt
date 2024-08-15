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
		if gw.BaudRate != baudRate || gw.Timeout != timeout {
			// (23/07/2024 kontornl) may cause memory leak without deleting, need inspection
			err := gw.cli.Close()
			time.Sleep(100 * time.Millisecond)
			if err != nil {
				return err
			}
		}
	}
	gw.BaudRate = baudRate
	gw.Timeout = timeout
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
		return
	}
	gw.cli = cli
	return
}

func (gw *MBRTGateway) Reconnect() (err error) {
	for retry := 0; retry <= 5; retry++ {
		gw.cli.Close()
		time.Sleep(time.Duration(retry) * 50 * time.Millisecond)
		err = gw.cli.Open()
		if err == nil {
			break
		}
	}
	return
}

func (gw *MBRTGateway) GetClient() (cli *modbus.ModbusClient) {
	cli = gw.cli
	return
}

type MBRTGateway struct {
	cli      *modbus.ModbusClient
	BaudRate uint
	Timeout  time.Duration
}

type IMBRTGateway interface {
	Init(netAddr string, baudRate uint, timeout time.Duration) (err error)
	Reconnect() (err error)
	GetClient() (cli *modbus.ModbusClient)
}
