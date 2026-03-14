Require dtkt

Set Shell "bash"
Set FontFamily "IBM Plex Mono"
Set FontSize 16
Set Width 1000
Set Height 600

Type `cat ./example-echo-flow/inputs.jsonl | jq -c` Enter@500ms

Sleep 1s

Type `cat ./example-echo-flow/inputs.jsonl | jq -c \` Enter@500ms
Type ` | dtkt flow run ./example-echo-flow/flow.dtkt.yaml` Enter@500ms

Sleep 5s
