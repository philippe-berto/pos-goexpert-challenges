mocks:
	rm -rf mocks
	go install go.uber.org/mock/mockgen@latest
	go generate -tags=tool mockgen.go

test:
	go fmt ./...
	go test -count 1 -vet all ./...