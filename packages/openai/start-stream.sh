#!/bin/bash

echo "Starting stream"
dtkt pkg base-service stream-pull-events -s localhost:9090 -p OpenAI@0.1.0 -c test-data/config.json -f test-data/openai_stream_events.json | jq -c '.payload' | go run ./cmd/audio "$1"
