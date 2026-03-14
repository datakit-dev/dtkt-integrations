Require dtkt

Set Shell "bash"
Set FontFamily "IBM Plex Mono"
Set FontSize 16
Set Width 1000
Set Height 600

Type `dtkt intgr create Example@0.1.0` Enter@500ms

Sleep 1s
Enter@500ms # identity
Type `An example integration.` Enter@500ms # description
Enter@500ms # icon
Enter@500ms # type

# runtimes
Down@500ms
Space@500ms
Down@500ms
Space@500ms
Enter@500ms

Enter@500ms # platforms

Sleep 1s
Enter@500ms # Services

Sleep 1s
Ctrl+Z@500ms

Type `dtkt call Greet` Enter@500ms
Sleep 1

Type `dtkt call Greet -d '{"name": "World"}'` Enter@500ms
Sleep 1

Type `dtkt call Greet -e` Enter@500ms
Sleep 1
Type `World` Enter@500ms
Sleep 1

Sleep 5s

Type `fg` Enter@500ms
Ctrl+C@500ms
