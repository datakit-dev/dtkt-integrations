#!/bin/bash

echo "Starting chat"
while true; do
  read -p "Enter text (type ctrl-c or 'exit' to quit): " input

  if [ "$input" == "exit" ]; then
    break
  fi

  payload="$(jq --arg text "$input" -c '.input.event.item.content[0].text = $text' test-data/openai_create_conversation_item.json)"

  echo "$payload" | dtkt pkg base-service exec-custom-action -s localhost:9090 -p OpenAI@0.1.0 -c test-data/config.json -f - | jq '.output'
  dtkt pkg base-service exec-custom-action -s localhost:9090 -p OpenAI@0.1.0 -c test-data/config.json -f test-data/openai_response_create.json | jq '.output'
done
