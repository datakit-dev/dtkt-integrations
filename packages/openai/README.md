# OpenAI Integration

## Realtime API

To test text & audio streaming, follow these steps:

1. Build [DataKit CLI](https://github.com/datakit-dev/dtkt-cli) and add it somewhere accessible to your $PATH:

```shell
  cd path/to/dtkt-cli
  task build
  task install
  which dtkt
  # Add $HOME/.local/bin to your $PATH if it is not found
```

2. Change to this directory:

```shell
  cd path/to/dtkt-integrations/packages/openai
```

3. Add OpenAI api key to `test-data/config.json` with the following format:

```json
{
  "api_key": "..."
}
```

4. Run the following commands:

```shell
  # Start package in dev mode:
  dtkt intgr dev

  # Start the Realtime API stream script:
  ./start-stream.sh # Optionally provide an output file (must have .pcm extension) to save audio stream (e.g. $HOME/Desktop/openai_audio.pcm)

  # Start the Chat input loop script:
  ./start-chat.sh
```
