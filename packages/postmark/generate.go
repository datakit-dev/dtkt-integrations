package main

//go:generate go tool github.com/ogen-go/ogen/cmd/ogen --target pkg/serveroapi/gen -package serveroapigen --clean pkg/serveroapi/openapi.json
//go:generate go tool github.com/ogen-go/ogen/cmd/ogen --target pkg/accountoapi/gen -package accountoapigen --clean pkg/accountoapi/openapi.json
