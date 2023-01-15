@echo off
:: The call source is from proto
:: liqi.json version v0.10.180.w
protoc --go_out=. --go-grpc_out=. proto/*.proto