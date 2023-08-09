# Majsoul Go

Majsoul Go is a library written in Go that provides an interface for interacting with Majsoul, an online mahjong game
server.

## Installation

Install Majsoul Go using the following command:

```
go get github.com/constellation39/majsoul
```

## Subpackages

The Majsoul Go project includes the following subpackages:

- **proto** and **tools**: These subpackages are used during development. They include the process of
  converting `liqi.json` to Go code using `.proto` and gRPC.
- **message**: This subpackage is generated from `proto` and `tools` and contains the code for handling messages.
- **logger**: Provides logging functionality. It uses the `zap` library for logging and offers logging configuration in
  both development and production modes.
- **utils**: Provides some utility functions, including password hashing, message decoding, and UUID generation.
- **network**: Contains network-related code.

## Usage Example

Here is a simple example of how to use this library to interact with the Majsoul server:

```go
package main

import (
  "context"
  "fmt"
  "github.com/constellation39/majsoul"
  "github.com/constellation39/majsoul/message"
  "os"
  "time"
)

func main() {
  account, exists := os.LookupEnv("account")
  if !exists {
    panic("account is required.")
  }

  password, exists := os.LookupEnv("password")
  if !exists {
    panic("password is required.")
  }

  majSoul := majsoul.NewMajSoul(&majsoul.Config{ProxyAddress: ""})
  { 
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()
    err := majSoul.LookupGateway(ctx, majsoul.ServerAddressList)
    if err != nil {
      panic(err)
    }
  }

  {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()
    resLogin, err := majSoul.Login(ctx, account, password)
    if err != nil {
      panic(err)
    }
    if resLogin.Error != nil {
      panic(resLogin.Error)
    }
  }

  {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()
    friendList, err := majSoul.LobbyClient.FetchFriendList(ctx, &message.ReqCommon{})
    if err != nil {
      panic(err)
    }
    fmt.Printf("%v", friendList)
  }
}

```
[example](https://github.com/constellation39/majsoul/tree/master/example)
