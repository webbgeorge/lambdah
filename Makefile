test:
	bash -c 'diff -u <(echo -n) <(gofmt -s -d .)'
	go vet ./...
	gosec ./...
	nancy go.sum
	go test ./... -race -v -coverprofile=coverage.out

install-dev-dependencies-mac:
	go get -u github.com/securego/gosec/cmd/gosec
	go get -u github.com/mattn/goveralls
	curl -LO https://github.com/sonatype-nexus-community/nancy/releases/download/v0.3.1/nancy-darwin.amd64-v0.3.1 && mv nancy-darwin.amd64-v0.3.1 /usr/local/bin/nancy && chmod +x /usr/local/bin/nancy

install-dev-dependencies-linux:
	go get -u github.com/securego/gosec/cmd/gosec
	go get -u github.com/mattn/goveralls
	curl -LO https://github.com/sonatype-nexus-community/nancy/releases/download/v0.3.1/nancy-linux.amd64-v0.3.1 && mv nancy-linux.amd64-v0.3.1 /usr/local/bin/nancy && chmod +x /usr/local/bin/nancy

.PHONY: test
