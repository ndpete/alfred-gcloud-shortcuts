TARGET_DIR := target
WORKFLOW_FILE := $(TARGET_DIR)/alfred-gcloud-shortcuts.alfredworkflow
LIPO ?= $(shell command -v lipo 2>/dev/null || command -v $(shell go env GOPATH 2>/dev/null)/bin/lipo 2>/dev/null)

.PHONY: all clean build build-projects build-products workflow install sort-products

all: build

clean:
	@rm -rf $(TARGET_DIR) bin

$(TARGET_DIR):
	@mkdir -p $(TARGET_DIR)

build-projects:
	@mkdir -p bin
ifeq ($(LIPO),)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/projects cmd/projects/*.go
else
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/projects-arm64 cmd/projects/*.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/projects-amd64 cmd/projects/*.go
	$(LIPO) -create -output bin/projects bin/projects-arm64 bin/projects-amd64
	@rm -f bin/projects-arm64 bin/projects-amd64
endif

build-products:
	@mkdir -p bin
ifeq ($(LIPO),)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/products cmd/products/*.go
else
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/products-arm64 cmd/products/*.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/products-amd64 cmd/products/*.go
	$(LIPO) -create -output bin/products bin/products-arm64 bin/products-amd64
	@rm -f bin/products-arm64 bin/products-amd64
endif

build: build-projects build-products

workflow: build | $(TARGET_DIR)
	@rm -f $(WORKFLOW_FILE)
	zip $(WORKFLOW_FILE) \
		info.plist \
		icon.png \
		products.json \
		bin/products \
		bin/projects

install: workflow
	open $(WORKFLOW_FILE)

sort-products:
	cat products.json | jq -s '.[] | sort_by(.name)' > products_sorted.json
	cp products_sorted.json products.json
	rm products_sorted.json
