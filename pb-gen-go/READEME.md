# pb-gen-go

An example usage of generating proto to golang by using buf.

## Buf

Document: https://buf.build/docs/introduction

Buf builds tooling to make schema-driven, Protobuf-based API development reliable and user-friendly for service producers and consumers.

## Dependencies
### buf
For Mac user, install by homebrew
```shell
 brew install bufbuild/buf/buf
```
### protoc-gen tools
```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.29.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
```
## Usage
### 1. Write proto file
Placed in `proto` package

### 2. lint

Run the following commands
```shell
buf lint
```
```shell
buf format --diff --exit-code
```

### 3. Generate
Run

```shell
go generate ./...
```