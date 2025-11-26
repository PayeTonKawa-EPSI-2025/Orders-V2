# Variables
BINARY=build/paye-ton-kawa--orders
DOCKER_IMAGE=ghcr.io/payetonkawa-epsi-2025/orders-v2/paye-ton-kawa--orders
SRC=cmd/main.go

build: $(BINARY)

$(BINARY): $(SRC)
	@mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o $@ $<

build-image: build
	@if [ -z "$(VERSION)" ]; then \
	  echo "Error: VERSION variable is required. Use: make build-image VERSION=<tag>" >&2; \
	  exit 1; \
	fi

	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

clean:
	rm -rf build
