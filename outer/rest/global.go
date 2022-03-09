package rest

type EnableSvrMsg struct {
	Enable bool
	Ext    bool
	Port   uint16
}

var EnableSvrChan = make(chan EnableSvrMsg, 10)
