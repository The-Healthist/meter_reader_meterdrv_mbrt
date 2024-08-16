package gateway

import (
	"net"
	"os"
	"sync"
	"syscall"
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

func (gw *MBRTGateway) Reinit() (err error) {
	err = gw.Init(gw.netAddr, gw.BaudRate, gw.Timeout)
	return
}
func (gw *MBRTGateway) Init(netAddr string, baudRate uint, timeout time.Duration) (err error) {
	var cli *modbus.ModbusClient
	if gw.cli != nil {
		if gw.BaudRate != baudRate || gw.Timeout != timeout {
			// (23/07/2024 kontornl) may cause memory leak without deleting, need inspection
			// (16/08/2024 kontornl) close without checking error after it
			// willing to reopen no matter what happened here ,especially errNetClosing
			gw.cli.Close()
			time.Sleep(100 * time.Millisecond)
		}
	}
	gw.netAddr = netAddr
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
	if gw.LastErr != nil {
		if assertedErr, ok := gw.LastErr.(*net.OpError); ok {
			if assertedErr, ok := assertedErr.Err.(*os.SyscallError); ok {
				if errNo, ok := assertedErr.Err.(syscall.Errno); ok {
					if errNo == syscall.ECONNREFUSED || errNo == 0x274d /* WSAECONNREFUSED */ {
						time.Sleep(5 * time.Second)
					}
				}
			}
		}
	}
	err = cli.Open()
	gw.LastErr = err
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
		// check lasterr
		err = gw.cli.Open()
		if err == nil {
			break
		}
	}
	gw.LastErr = err
	return
}

func (gw *MBRTGateway) GetClient() (cli *modbus.ModbusClient) {
	cli = gw.cli
	return
}

func (gw *MBRTGateway) GetLock() (mtx *sync.RWMutex) {
	mtx = &gw.mtx
	return
}

type MBRTGateway struct {
	cli      *modbus.ModbusClient
	BaudRate uint
	netAddr  string
	Timeout  time.Duration
	mtx      sync.RWMutex
	LastErr  error
}

type IMBRTGateway interface {
	Init(netAddr string, baudRate uint, timeout time.Duration) (err error)
	Reconnect() (err error)
	GetClient() (cli *modbus.ModbusClient)
}
