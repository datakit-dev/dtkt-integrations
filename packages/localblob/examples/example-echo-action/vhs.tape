Require dtkt

Set Shell "bash"
Set FontFamily "IBM Plex Mono"
Set FontSize 16
Set Width 1000
Set Height 600

Type `cat examples/example-echo-action/request.json` Enter@500ms

Sleep 1s

Type `dtkt action execute-action \` Enter@500ms
Type `  --conn localblob \` Enter@500ms
Type `  --name actions/example-echo \` Enter@500ms
Type `  -f examples/example-echo-action/request.json` Enter@500ms

Sleep 5s
