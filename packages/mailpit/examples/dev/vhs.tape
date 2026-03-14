Require dtkt

Set Width 1000
Set Height 600

Set Shell "bash"

Set Theme {"name":"DataKit","black":"#060818","red":"#fb7185","green":"#3fea83","yellow":"#fdba74","blue":"#a2bdff","magenta":"#d8b4fe","cyan":"#a2bdff","white":"#efefef","brightBlack":"#989898","brightRed":"#fb7185","brightGreen":"#3fea83","brightYellow":"#fdba74","brightBlue":"#a2bdff","brightMagenta":"#d8b4fe","brightCyan":"#a2bdff","brightWhite":"#efefef","background":"#060818","foreground":"#efefef","selection":"#2a3182","cursor":"#3fea83"}
Set Margin 20
Set MarginFill "#2a3182"
Set BorderRadius 10

Hide
Type `dtkt intgr delete` Enter
Type `sh -c 'sleep 10 && touch example-file.txt && sleep 3 && rm example-file.txt' &` Enter
Sleep 100ms
Type `clear` Enter
Sleep 1s
Show

Type `dtkt intgr dev` Enter@1s

Sleep 5s

Enter

Sleep 5s

# File created

Sleep 3s

# File deleted

Sleep 3s

Ctrl+C

Sleep 2s
