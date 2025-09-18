GOOS ?= darwin
GOARCH ?= arm64
BIN_DIR := Workflow
CHAT_OUT := $(BIN_DIR)/chatgpt
DALLE_OUT := $(BIN_DIR)/dalle

.PHONY: build chat dalle clean

build: chat dalle

chat: | $(BIN_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(CHAT_OUT) ./cmd/chatgpt

dalle: | $(BIN_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(DALLE_OUT) ./cmd/dalle

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

clean:
	rm -f $(CHAT_OUT) $(DALLE_OUT)

