package pub

type MemStorStru struct {
	Ips map[string]struct{} `json:"ips"`
}

var MemStor MemStorStru
