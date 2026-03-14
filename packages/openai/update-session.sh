#!/bin/bash

echo "Updating session"
dtkt pkg base-service exec-custom-action -s localhost:9090 -p OpenAI@0.1.0 -c test-data/config.json -f test-data/openai_update_session.json | jq '.output'
