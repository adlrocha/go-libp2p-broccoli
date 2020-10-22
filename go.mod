module github.com/libp2p/go-libp2p-broccoli

go 1.14

require (
	github.com/dropbox/rust-brotli v0.0.0-20200819074738-458f5780a376
	github.com/libp2p/go-libp2p-core v0.0.0-00010101000000-000000000000
	go.uber.org/multierr v1.6.0
)

replace github.com/libp2p/go-libp2p-core => github.com/adlrocha/go-libp2p-core v0.6.2-0.20201008055055-b309947fc237
