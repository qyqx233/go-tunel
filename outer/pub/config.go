package pub

type MemStorStru struct {
	Ips map[string]struct{} `json:"ips"`
}

var MemStor MemStorStru

func init() {
	MemStor.Ips = make(map[string]struct{})
}
