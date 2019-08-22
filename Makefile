SRCS := $(shell find . -type f -name '*.go')
REVISION := `git show --date=format:'%Y/%m/%d_%H:%M:%S' --quiet --pretty=format:"%cd_%H" HEAD`
LD_FLAGS=-ldflags "-X main.Revision=$(REVISION)"
OUT_DIR := out
APP := $(OUT_DIR)/app

all: $(APP)

$(APP): $(SRCS)
	docker run -i --rm \
	-v `pwd`:/work \
	-v `pwd`/mod:/go/pkg/mod \
	-w /work \
	-e GOOS=linux \
	-e GOARCH=amd64 \
	-e CGO_ENABLED=0 \
	golang:latest \
	go build $(LD_FLAGS) -o $(APP)

run: all
	docker run -it --rm \
	-v `pwd`/$(OUT_DIR):/appdir \
	-w /appdir \
	-p 8080:80 \
	alpine ./app

clean:
	rm -f $(APP)
	docker run -it --rm \
	-v `pwd`/mod:/go/pkg/mod \
	golang:latest go clean -modcache 1>/dev/null || :

.PHONY: all run clean