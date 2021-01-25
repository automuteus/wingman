module github.com/automuteus/wingman

go 1.15

require (
	github.com/automuteus/galactus v1.2.2
	github.com/automuteus/utils v0.0.10
	github.com/googollee/go-socket.io v1.4.4
	github.com/gorilla/mux v1.8.0
	go.uber.org/zap v1.16.0
)

// TODO replace when V7 comes out
replace (
	github.com/automuteus/galactus v1.2.2 => github.com/automuteus/galactus v1.2.3-0.20210125064638-a00060928562
	github.com/automuteus/utils v0.0.10 => github.com/automuteus/utils v0.0.11-0.20210117090606-d48d8a0c6a4b
)
