package pci

type DeviceClass uint16

const (
	Undefined                  DeviceClass = 0x0000
	Undefined_VGA              DeviceClass = 0x0001
	Storage_SCSI               DeviceClass = 0x0100
	Storage_IDE                DeviceClass = 0x0101
	Storage_Floppy             DeviceClass = 0x0102
	Storage_IPI                DeviceClass = 0x0103
	Storage_RAID               DeviceClass = 0x0104
	Storage_SATA               DeviceClass = 0x0106
	Storage_SAS                DeviceClass = 0x0107
	Storage_Other              DeviceClass = 0x0180
	Network_Ethernet           DeviceClass = 0x0200
	Network_Token_Ring         DeviceClass = 0x0201
	Network_FDDI               DeviceClass = 0x0202
	Network_ATM                DeviceClass = 0x0203
	Network_Other              DeviceClass = 0x0280
	Display_VGA                DeviceClass = 0x0300
	Display_XGA                DeviceClass = 0x0301
	Display_3D                 DeviceClass = 0x0302
	Display_Other              DeviceClass = 0x0380
	Multimedia_Video           DeviceClass = 0x0400
	Multimedia_Audio           DeviceClass = 0x0401
	Multimedia_Phone           DeviceClass = 0x0402
	Multimedia_Audio_Device    DeviceClass = 0x0403
	Multimedia_Other           DeviceClass = 0x0480
	Memory_RAM                 DeviceClass = 0x0500
	Memory_Flash               DeviceClass = 0x0501
	Memory_Other               DeviceClass = 0x0580
	Bridge_Host                DeviceClass = 0x0600
	Bridge_ISA                 DeviceClass = 0x0601
	Bridge_EISA                DeviceClass = 0x0602
	Bridge_MC                  DeviceClass = 0x0603
	Bridge_PCI                 DeviceClass = 0x0604
	Bridge_PCMCIA              DeviceClass = 0x0605
	Bridge_NUBUS               DeviceClass = 0x0606
	Bridge_CARDBUS             DeviceClass = 0x0607
	Bridge_RACEWAY             DeviceClass = 0x0608
	Bridge_Other               DeviceClass = 0x0680
	Communication_Serial       DeviceClass = 0x0700
	Communication_Parallel     DeviceClass = 0x0701
	Communication_Multi_Serial DeviceClass = 0x0702
	Communication_Modem        DeviceClass = 0x0703
	Communication_Other        DeviceClass = 0x0780
	System_PIC                 DeviceClass = 0x0800
	System_DMA                 DeviceClass = 0x0801
	System_Timer               DeviceClass = 0x0802
	System_RTC                 DeviceClass = 0x0803
	System_PCI_Hotplug         DeviceClass = 0x0804
	System_SDHCI               DeviceClass = 0x0805
	System_Other               DeviceClass = 0x0880
	Input_Keyboard             DeviceClass = 0x0900
	Input_Pen                  DeviceClass = 0x0901
	Input_Mouse                DeviceClass = 0x0902
	Input_Scanner              DeviceClass = 0x0903
	Input_GAMEPORT             DeviceClass = 0x0904
	Input_Other                DeviceClass = 0x0980
	Docking_GENERIC            DeviceClass = 0x0a00
	Docking_Other              DeviceClass = 0x0a80
	Processor_386              DeviceClass = 0x0b00
	Processor_486              DeviceClass = 0x0b01
	Processor_Pentium          DeviceClass = 0x0b02
	Processor_Alpha            DeviceClass = 0x0b10
	Processor_Powerpc          DeviceClass = 0x0b20
	Processor_MIPS             DeviceClass = 0x0b30
	Processor_CO               DeviceClass = 0x0b40
	Serial_Firewire            DeviceClass = 0x0c00
	Serial_Access              DeviceClass = 0x0c01
	Serial_SSA                 DeviceClass = 0x0c02
	Serial_USB                 DeviceClass = 0x0c03
	Serial_Fiber               DeviceClass = 0x0c04
	Serial_SMBUS               DeviceClass = 0x0c05
	Wireless_RF_Controller     DeviceClass = 0x0d10
	Intelligent_I2O            DeviceClass = 0x0e00
	Satellite_TV               DeviceClass = 0x0f00
	Satellite_Audio            DeviceClass = 0x0f01
	Satellite_Voice            DeviceClass = 0x0f03
	Satellite_Data             DeviceClass = 0x0f04
	Crypt_Network              DeviceClass = 0x1000
	Crypt_entertainment        DeviceClass = 0x1001
	Crypt_Other                DeviceClass = 0x1080
	SP_DPIO                    DeviceClass = 0x1100
	SP_Other                   DeviceClass = 0x1180
)

const (
	Broadcom VendorID = 0x14e4
	Intel    VendorID = 0x8086
)

//go:generate stringer -type=DeviceClass,VendorID
