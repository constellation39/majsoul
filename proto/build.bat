@echo off
:: The call source is from proto
:: liqi.json version v0.10.217.w
:: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
protoc --go_out=plugins=grpc:. --go_opt=Mliqi.proto=../message liqi.proto