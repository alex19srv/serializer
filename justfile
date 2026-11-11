set shell := ["bash", "-c"]

build-codegen:
  go build -o ./target/codegen ./cmd/codegen

generate: build-codegen
  ./target/codegen ./tests/structs.go ./tests/structs_gen.go

test:
  go test -v ./tests/...
