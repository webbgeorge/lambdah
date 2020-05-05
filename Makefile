test:
	bash -c 'diff -u <(echo -n) <(gofmt -s -d .)'
	go vet ./...
	go test ./... -v -covermode=atomic -coverprofile=coverage.out

.PHONY: test
