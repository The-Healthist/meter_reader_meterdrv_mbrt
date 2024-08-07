package powermeter

// (23/07/2024 kontornl) may use const here, need inspection
// register metadata of DDS4921 oredered by item ID, such as ID_VOLTAGE
var regMetaDDS4921 = []RegMeta{
	{
		regAddr:   0x0000,
		length:    1,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.1,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		regAddr:   0x0003,
		length:    1,
		readable:  true,
		writable:  false,
		hasSymbol: true,
		override:  0.01,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		regAddr:   0x0007,
		length:    1,
		readable:  true,
		writable:  false,
		hasSymbol: true,
		override:  1,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		regAddr:   0x000B,
		length:    1,
		readable:  true,
		writable:  false,
		hasSymbol: true,
		override:  1,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		regAddr:   0x000F,
		length:    1,
		readable:  true,
		writable:  false,
		hasSymbol: true,
		override:  1,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		regAddr:   0x0013,
		length:    1,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.001,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		length: 0,
	},
	{
		regAddr:   0x001A,
		length:    1,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.01,
	},
	{
		regAddr:   0x001D,
		length:    2,
		readable:  true,
		writable:  false,
		hasSymbol: true,
		override:  0.01,
	},
	{
		regAddr:   0x0027,
		length:    2,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.01,
	},
	{
		regAddr:   0x0031,
		length:    2,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.01,
	},
	{
		regAddr:   0x003B,
		length:    2,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.01,
	},
	{
		regAddr:   0x0045,
		length:    2,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.01,
	},
	{
		regAddr:   0x004F,
		length:    2,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.01,
	},
	{
		regAddr:   0x0061,
		length:    1,
		readable:  true,
		writable:  true,
		hasSymbol: false,
		override:  0.01,
	},
	{
		length: 0,
	},
}

var switchMetaDDS4921 = []SwitchMeta{
	{
		ctlAddr:        0x0010,
		ctlTripCmd:     0xAAAA,
		ctlCloseCmd:    0x5555,
		statusAddr:     0x0064,
		statusTripVal:  0x00AA,
		statusCloseVal: 0x0055,
	},
}
