.PHONY: clean test

sakura-secrets-cli: go.* *.go
	go build -o $@ ./cmd/sakura-secrets-cli

clean:
	rm -rf sakura-secrets-cli dist/

test:
	go test -v ./...

install:
	go install github.com/fujiwara/sakura-secrets-cli/cmd/sakura-secrets-cli

dist:
	goreleaser build --snapshot --clean
