# ElevenLabs Go Scripts

Two standalone Go scripts are included:

- `cmd/tts/main.go`: generate speech from text with ElevenLabs TTS
- `cmd/stt/main.go`: transcribe a local audio file or public audio URL with ElevenLabs STT

## Requirements

- Go installed
- An ElevenLabs account
- An ElevenLabs API key
- A valid ElevenLabs voice for TTS

## ElevenLabs Portal Setup

1. Sign in to ElevenLabs.
2. Create an API key in the developer/settings area.
3. Pick an existing voice or create your own voice.
4. Copy the voice ID for that voice.
5. Make sure your account has available credits.

## Environment File

Create a `.env` file in the project root:

```env
API_KEY=your_elevenlabs_api_key
VOICE_ID=your_voice_id
```

These names match the current scripts exactly.

## Run Directly With Go

TTS example:

```bash
go run cmd/tts/main.go -text "Hello from ElevenLabs" -out "speech.mp3"
```

STT example with local file:

```bash
go run cmd/stt/main.go -in "audio.wav"
```

STT example with public URL:

```bash
go run cmd/stt/main.go -url "https://example.com/audio.ogg"
```

## Script Flags

### `cmd/tts/main.go`

```bash
go run cmd/tts/main.go \
  -text "Hello from ElevenLabs" \
  -out "speech.mp3" \
  -model-id "eleven_multilingual_v2" \
  -format "mp3_44100_128"
```

Available flags:

- `-text`: required text to synthesize
- `-out`: output audio file path, default `output.mp3`
- `-model-id`: TTS model, default `eleven_multilingual_v2`
- `-format`: output format, default `mp3_44100_128`
- `-base-url`: API base URL, default `https://api.elevenlabs.io`

### `cmd/stt/main.go`

```bash
go run cmd/stt/main.go \
  -in "audio.wav" \
  -model-id "scribe_v2"
```

Or pass a public URL directly to ElevenLabs without downloading it locally:

```bash
go run cmd/stt/main.go \
  -url "https://example.com/audio.ogg" \
  -model-id "scribe_v2"
```

Available flags:

- `-in`: local input audio/video file
- `-url`: public HTTPS audio/video URL sent directly to ElevenLabs as `cloud_storage_url`
- `-model-id`: STT model, default `scribe_v2`
- `-raw`: print raw JSON instead of only transcript text
- `-base-url`: API base URL, default `https://api.elevenlabs.io`

Provide exactly one of `-in` or `-url`.

## Use Make

Show available commands:

```bash
make help
```

Run TTS:

```bash
make tts TEXT="Hello from ElevenLabs" OUT="speech.mp3"
```

Run STT:

```bash
make stt IN="audio.wav"
```

Run STT with a public URL:

```bash
make stt URL="https://example.com/audio.ogg"
```

Build binaries:

```bash
make build
```

This creates:

- `bin/tts`
- `bin/stt`

Run built binaries:

```bash
./bin/tts -text "Hello from ElevenLabs" -out "speech.mp3"
./bin/stt -in "audio.wav"
./bin/stt -url "https://example.com/audio.ogg"
```

## Make Variables

- `TEXT`: text for TTS
- `OUT`: output file for TTS, default `output.mp3`
- `IN`: input file for STT
- `URL`: public HTTPS URL for STT
- `TTS_MODEL_ID`: default `eleven_multilingual_v2`
- `STT_MODEL_ID`: default `scribe_v2`
- `TTS_FORMAT`: default `mp3_44100_128`

Example:

```bash
make tts TEXT="Testing voice" OUT="test.mp3" TTS_MODEL_ID="eleven_multilingual_v2"
make stt IN="sample.wav" STT_MODEL_ID="scribe_v2"
make stt URL="https://example.com/audio.ogg" STT_MODEL_ID="scribe_v2"
```

## Notes

- TTS uses `VOICE_ID` from `.env`.
- STT does not use `VOICE_ID`; ElevenLabs STT only needs `API_KEY`, `model_id`, and either an uploaded file or a public HTTPS URL.
- If your API key was exposed, rotate it in the ElevenLabs portal.
