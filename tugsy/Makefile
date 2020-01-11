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
	montage \
		${SPRITES}/special/H0-S0-V50.png  \
		${SPRITES}/special/R1-H0.png \
		${SPRITES}/special/R2-H0.png \
		${SPRITES}/special/R3-H0.png \
		${SPRITES}/special/R4-H0.png \
		-tile 1x6 \
		-geometry 32x32x0x0 \
		-background transparent \
		${SPRITES}/special.png
