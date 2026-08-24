module github.com/checkpoint-restore/go-criu/v8

go 1.25.0

require (
	github.com/aperturerobotics/protobuf-go-lite v0.16.0
	github.com/pierrec/lz4/v4 v4.1.27
	github.com/spf13/cobra v1.10.2
	golang.org/x/sys v0.47.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

tool (
	github.com/aperturerobotics/protobuf-go-lite/cmd/protoc-gen-go-lite
	google.golang.org/protobuf/cmd/protoc-gen-go
)
