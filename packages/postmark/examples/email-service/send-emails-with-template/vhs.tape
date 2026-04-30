Require dtkt

Set Width 1000
Set Height 600

Set Shell "bash"

Set Theme {"name":"DataKit","black":"#060818","red":"#fb7185","green":"#3fea83","yellow":"#fdba74","blue":"#a2bdff","magenta":"#d8b4fe","cyan":"#a2bdff","white":"#efefef","brightBlack":"#989898","brightRed":"#fb7185","brightGreen":"#3fea83","brightYellow":"#fdba74","brightBlue":"#a2bdff","brightMagenta":"#d8b4fe","brightCyan":"#a2bdff","brightWhite":"#efefef","background":"#060818","foreground":"#efefef","selection":"#2a3182","cursor":"#3fea83"}
Set Margin 20
Set MarginFill "#2a3182"
Set BorderRadius 10

Hide
Type `dtkt intgr start --detach` Enter
Type `dtkt connect delete postmark` Enter
Sleep 100ms
Type `dtkt create connection postmark --intgr postmark -f examples/configs/postmark.json --edit=false` Enter
Sleep 100ms
Type `clear` Enter
Sleep 1s
Show

Type `dtkt call SendEmailsWithTemplate \` Enter@500ms
Type `  --conn postmark \` Enter@500ms
Type `  --edit false \` Enter@500ms
Type `  -d '{` Enter
Type `  }'` Enter@1s

Sleep 5s
