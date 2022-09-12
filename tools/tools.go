//go:build tools
// +build tools

package tools

import (
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
)

// Get https://game.maj-soul.net/1/v0.10.135.w/res/proto/liqi.json
