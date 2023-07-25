xpmt: clean
	CGO_ENABLED=1 go build -o xpmt cmd/xpmt/main.go

docker:
	docker build -t experiment . --progress plain
	docker run -it --rm --name experiment-run experiment

release: copy-lib-release xpmt

debug: copy-lib-debug xpmt

clean:
	rm -f xpmt

all: darwin linux

darwin: darwin-amd64 darwin-arm64

linux: linux-amd64 linux-arm64

darwin-amd64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o cmd/xpmt/bin/darwin/amd64/xpmt cmd/xpmt/main.go

darwin-arm64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o cmd/xpmt/bin/darwin/arm64/xpmt cmd/xpmt/main.go

linux-amd64:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o cmd/xpmt/bin/linux/amd64/xpmt cmd/xpmt/main.go

linux-arm64:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o cmd/xpmt/bin/linux/arm64/xpmt cmd/xpmt/main.go

# expects experiment-evaluation lives in same directory and experiment-go-server
copy-lib-debug:
	# macosX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/debugStatic/libevaluation_interop_api.h internal/evaluation/lib/macosX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/debugStatic/libevaluation_interop.a internal/evaluation/lib/macosX64/
	# macosArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/debugStatic/libevaluation_interop_api.h internal/evaluation/lib/macosArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/debugStatic/libevaluation_interop.a internal/evaluation/lib/macosArm64/
	# linuxX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/debugStatic/libevaluation_interop_api.h internal/evaluation/lib/linuxX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/debugStatic/libevaluation_interop.a internal/evaluation/lib/linuxX64/
	# linuxArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/debugStatic/libevaluation_interop_api.h internal/evaluation/lib/linuxArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/debugStatic/libevaluation_interop.a internal/evaluation/lib/linuxArm64/


# expects experiment-evaluation lives in same directory and experiment-go-server
copy-lib-release:
	# macosX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/releaseStatic/libevaluation_interop_api.h internal/evaluation/lib/macosX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosX64/releaseStatic/libevaluation_interop.a internal/evaluation/lib/macosX64/
	# macosArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/releaseStatic/libevaluation_interop_api.h internal/evaluation/lib/macosArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/macosArm64/releaseStatic/libevaluation_interop.a internal/evaluation/lib/macosArm64/
	# linuxX64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/releaseStatic/libevaluation_interop_api.h internal/evaluation/lib/linuxX64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxX64/releaseStatic/libevaluation_interop.a internal/evaluation/lib/linuxX64/
	# linuxArm64
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/releaseStatic/libevaluation_interop_api.h internal/evaluation/lib/linuxArm64/
	cp ../experiment-evaluation/evaluation-interop/build/bin/linuxArm64/releaseStatic/libevaluation_interop.a internal/evaluation/lib/linuxArm64/
