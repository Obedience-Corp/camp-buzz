# camp-buzz post — footer contract + fake buzz acceptance
#
# Record: just vhs post

Output docs/demos/camp-buzz-post.gif

Set Shell "bash"
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 15
Set Width 1000
Set Height 560
Set Padding 16
Set Margin 0
Set BorderRadius 0
Set Framerate 24
Set TypingSpeed 32ms
Set Theme {"name":"termcast","background":"#0d0d11","foreground":"#f8f8f2","cursor":"#F2721C","selection":"#2a2a33","black":"#16161a","red":"#ff5f57","green":"#50fa7b","yellow":"#febc2e","blue":"#57a6ff","magenta":"#bd93f9","cyan":"#8be9fd","white":"#f8f8f2","brightBlack":"#6b7280","brightRed":"#ff6e67","brightGreen":"#69ff94","brightYellow":"#ffd23f","brightBlue":"#7ab8ff","brightMagenta":"#d0a9ff","brightCyan":"#a4f0ff","brightWhite":"#ffffff"}

Hide
Type "export CAMP_BUZZ_VHS_ROOT=/tmp/camp-buzz-vhs-post" Enter
Type "rm -rf $CAMP_BUZZ_VHS_ROOT" Enter
Type "FIX=$(bash scripts/vhs-fixture.sh)" Enter
Type "source $FIX/env.sh" Enter
Type "camp-buzz bind --channel 22222222-2222-4222-8222-222222222222 --relay http://127.0.0.1:3000 --festival CI0009 >/dev/null" Enter
Type "clear" Enter
Show

Type "camp-buzz doctor" Enter
Sleep 1400ms
Type "clear" Enter
Sleep 300ms

Type "camp-buzz post -m 'Gate passed: quality checks green' --task FEST-a1b2c3 --path festivals/active/demo --gate pass" Enter
Sleep 1200ms
Type "clear" Enter
Sleep 300ms

Type "cat $BUZZ_FAKE_LOG" Enter
Sleep 2200ms
