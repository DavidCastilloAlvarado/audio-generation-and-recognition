.PHONY: help fmt build tts stt clean

GO ?= go
BIN_DIR ?= bin
TTS_BIN := $(BIN_DIR)/tts
STT_BIN := $(BIN_DIR)/stt

TEXT ?=
OUT ?= output.mp3
IN ?=
URL ?=
TTS_MODEL_ID ?= eleven_multilingual_v2
STT_MODEL_ID ?= scribe_v2
TTS_FORMAT ?= mp3_44100_128

help:
	@printf '%s\n' \
	  'Targets:' \
	  '  make help                         Show this help' \
	  '  make fmt                          Format Go files' \
	  '  make build                        Build bin/tts and bin/stt' \
	  '  make tts TEXT="hello"             Run TTS script' \
	  '  make stt IN="audio.wav"           Run STT with local file' \
	  '  make stt URL="https://..."        Run STT with public URL' \
	  '  make clean                        Remove built binaries'

fmt:
	$(GO) fmt ./...

build:
	mkdir -p "$(BIN_DIR)"
	$(GO) build -o "$(TTS_BIN)" cmd/tts/main.go
	$(GO) build -o "$(STT_BIN)" cmd/stt/main.go

tts:
	@if [ -z "$(TEXT)" ]; then printf '%s\n' 'TEXT is required. Example: make tts TEXT="Hello"'; exit 1; fi
	$(GO) run cmd/tts/main.go -text "$(TEXT)" -out "$(OUT)" -model-id "$(TTS_MODEL_ID)" -format "$(TTS_FORMAT)"

stt:
	@if [ -n "$(IN)" ] && [ -n "$(URL)" ]; then printf '%s\n' 'Provide only one of IN or URL.'; exit 1; fi
	@if [ -z "$(IN)" ] && [ -z "$(URL)" ]; then printf '%s\n' 'IN or URL is required. Example: make stt IN="audio.wav" or make stt URL="https://example.com/audio.ogg"'; exit 1; fi
	$(GO) run cmd/stt/main.go $(if $(IN),-in "$(IN)") $(if $(URL),-url "$(URL)") -model-id "$(STT_MODEL_ID)"

clean:
	rm -rf "$(BIN_DIR)"
