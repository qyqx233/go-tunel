package rest

//go:generate msgp

type TransportPdb struct {
	Enable bool
	Export bool
	Usage  string
}
