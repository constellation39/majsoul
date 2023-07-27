//go:build ignore
// +build ignore

// Package tools provides support for protocol buffers using the Go programming language.
//
// This package imports two other packages that are required for protocol buffer support.
package tools

import (
	// Package protoc-gen-go-grpc is imported for generating gRPC service code.
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"

	// Package protoc-gen-go is imported for generating protocol buffer code.
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
