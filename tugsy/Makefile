BUILD_OS := darwin linux
ARCH := amd64

SPRITES := Resources/sprites

build: clean $(BUILD_OS)

$(BUILD_OS):
	@for f in ./cmd/tugsy ; do \
		GOOS=$@ GOARCH=$(ARCH) CGO_ENABLED=1 go build -a -ldflags '-w' -o ./$$(basename $$f) $$f ;\
		echo $$(basename $$f) ;\
	done

clean:
	rm tugsy

special-sprites:
	cd ${SPRITES}/special && \
	montage \
		`ls -1 *.png` \
		-tile 1x \
		-geometry 32x32x0x0 \
		-background transparent \
		../special.png

flag-sprites:
	cd ${SPRITES}/flags && \
	montage \
		`ls -1 *.png` \
		-tile 1x \
		-geometry 64x64x0x0 \
		-background transparent \
		../flags.png
