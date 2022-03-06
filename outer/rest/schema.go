package rest

type ListTransport struct {
	Data []Transport `json:"data"`
}

type Transport struct {
	Port       int    `json:"port"`
	TargetHost string `json:"targetHost"`
	TargetPort int    `json:"targetPort"`
	Enable     bool   `json:"enable"`
}
