package rest

type ListTransport struct {
	Data []Transport `json:"data"`
}

type Transport struct {
	Port       int    `json:"port,omitempty"`
	TargetHost string `json:"targetHost,omitempty"`
	TargetPort int    `json:"targetPort,omitempty"`
	Enable     bool   `json:"enable,omitempty"`
	Usage      string `json:"usage,omitempty"`
}
