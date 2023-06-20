default: build

build:
	go build -o ~/.terraform.d/plugins/terraform.local/local/ssh/0.1.0/darwin_amd64/terraform-provider-ssh

install: build
	go install -v ./...

# See https://golangci-lint.run/
lint:
	golangci-lint run

# Generate docs and copywrite headers
generate:
	go generate ./...
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=4 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: build install lint generate fmt test testacc
