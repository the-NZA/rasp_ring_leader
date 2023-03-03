# build for windows, darwin with amd64 and darwin with arm

WIN=w.exe
MAC_ARM=macarm
MAC_AMD64=macamd64


.default: build

build:
	go build -o bin/$(MAC_ARM) -v && \
	GOOS=windows GOARCH=amd64 go build -o bin/$(WIN) -v && \
	GOOS=darwin GOARCH=amd64 go build -o bin/$(MAC_AMD64) -v