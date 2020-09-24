package message

// Message types sent by server
const (
	AuthenticationOK       = 0x52
	CommandComplete        = 0x43
	CommandDataDescription = 0x54
	Data                   = 0x44
	DumpBlock              = 0x3d
	DumpHeader             = 0x40
	ErrorResponse          = 0x45
	LogMessage             = 0x4c
	ParameterStatus        = 0x53
	PrepareComplete        = 0x31
	ReadyForCommand        = 0x5a
	RestoreReady           = 0x2b
	ServerHandshake        = 0x76
	ServerKeyData          = 0x4b
)

// Message types sent by client
const (
	ClientHandshake   = 0x56
	DescribeStatement = 0x44
	Dump              = 0x3e
	Execute           = 0x45
	ExecuteScript     = 0x51
	Flush             = 0x48
	OptimisticExecute = 0x4f
	Prepare           = 0x50
	Restore           = 0x3c
	RestoreBlock      = 0x3d
	RestoreEOF        = 0x2e
	Sync              = 0x53
	Terminate         = 0x58
)
