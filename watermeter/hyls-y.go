package watermeter

// (23/07/2024 kontornl) may use const here, need inspection
// register metadata of HYLS-Y oredered by item ID, such as ID_VOLUME
var regMetaHYLSY = []RegMeta{
	{
		regAddr:   0x0000,
		length:    2,
		readable:  true,
		writable:  false,
		hasSymbol: false,
		override:  0.01,
	},
}

var valveMetaHYLSY = []ValveMeta{
	{
		ctlAddr:        0x0001,
		ctlRegType:     REGTYPE_COIL,
		ctlCloseCmd:    0x0000,
		ctlOpenCmd:     0x0001,
		statusAddr:     0x0001,
		statusRegType:  REGTYPE_COIL,
		statusCloseVal: 0x0000,
		statusOpenVal:  0x0001,
	},
}
