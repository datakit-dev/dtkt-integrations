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
Type `dtkt connect delete email` Enter
Sleep 100ms
Type `dtkt connect create email --intgr email -f examples/configs/mailpit.json --edit=false` Enter
Type `dtkt connect delete mailpit` Enter
Type `dtkt connect create mailpit --intgr mailpit -f ../mailpit/examples/configs/mailpit.json` Enter
Sleep 100ms
Type `clear` Enter
Sleep 1s
Show

Type `dtkt call SendEmail \` Enter@500ms
Type `  --conn email \` Enter@500ms
Type `  --edit false \` Enter@500ms
Type `  -d '{` Enter
Type `    "email": {` Enter
Type `      "from": {` Enter
Type `        "name": "DataKit",` Enter
Type `        "address": "hello@withdatakit.com"` Enter
Type `      },` Enter
Type `      "to": [` Enter
Type `        {` Enter
Type `          "name": "User",` Enter
Type `          "address": "user@withdatakit.com"` Enter
Type `        }` Enter
Type `      ],` Enter
Type `      "subject": "[Test] Hello from the DataKit Email integration",` Enter
Type `      "textBody": "This is a test email",` Enter
Type `      "htmlBody": "<p>This is a test email</p>"` Enter
Type `    }` Enter
Type `  }'` Enter@1s

Sleep 5s

Type `dtkt call ExecuteAction \` Enter@500ms
Type `  --conn mailpit \` Enter@500ms
Type `  -d '{` Enter
Type `    "name": "actions/get-message",` Enter
Type `    "input": {` Enter
Type `      "@type": "type.googleapis.com/google.protobuf.Struct",` Enter
Type `      "value": {` Enter
Type `        "ID": "latest"` Enter
Type `      }` Enter
Type `    }` Enter
Type `  }'` Enter@1s

Sleep 5s

Type `# Sent and delivered! 💌` Enter@500ms

Sleep 5s
