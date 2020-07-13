module gotunel

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	go.uber.org/zap v1.13.0
)

replace gotunel/server => ./server

replace gotunel/lib => ./lib
