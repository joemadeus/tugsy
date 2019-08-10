BUILD_OS := darwin linux
ARCH := amd64

GOBUILD_PATH ?= papertrail/gobuild

build: clean $(BUILD_OS)

$(BUILD_OS):
	@for f in ./cmd/tugsy ; do \
		GOOS=$@ GOARCH=$(ARCH) go build -a -ldflags '-w' -o ./$$(basename $$f) $$f ;\
		echo ./target/$@-$(ARCH)/$$(basename $$f) ;\
	done

clean:
	rm tugsy
