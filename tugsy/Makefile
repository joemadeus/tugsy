BUILD_OS := darwin linux
ARCH := amd64

build: clean $(BUILD_OS)

$(BUILD_OS):
	@for f in ./cmd/tugsy ; do \
		GOOS=$@ GOARCH=$(ARCH) CGO_ENABLED=1 go build -a -ldflags '-w' -o ./$$(basename $$f) $$f ;\
		echo $$(basename $$f) ;\
	done

clean:
	rm tugsy
