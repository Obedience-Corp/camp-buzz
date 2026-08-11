# camp-buzz plugin tour — doctor, bind, show, post, hook-install
#
# Isolated fixture + fake buzz CLI (no real relay required).
# Record: just vhs tour

Output docs/demos/camp-buzz-tour.gif

Set Shell "bash"
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 15
Set Width 1100
Set Height 640
Set Padding 16
Set Margin 0
Set BorderRadius 0
Set Framerate 24
Set TypingSpeed 35ms
Set Theme {"name":"termcast","background":"#0d0d11","foreground":"#f8f8f2","cursor":"#F2721C","selection":"#2a2a33","black":"#16161a","red":"#ff5f57","green":"#50fa7b","yellow":"#febc2e","blue":"#57a6ff","magenta":"#bd93f9","cyan":"#8be9fd","white":"#f8f8f2","brightBlack":"#6b7280","brightRed":"#ff6e67","brightGreen":"#69ff94","brightYellow":"#ffd23f","brightBlue":"#7ab8ff","brightMagenta":"#d0a9ff","brightCyan":"#a4f0ff","brightWhite":"#ffffff"}

Hide
Type "export CAMP_BUZZ_VHS_ROOT=/tmp/camp-buzz-vhs-tour" Enter
Type "rm -rf $CAMP_BUZZ_VHS_ROOT" Enter
Type "FIX=$(bash scripts/vhs-fixture.sh)" Enter
Type "export CAMP_BUZZ_VHS_ROOT=$FIX" Enter
Type "source $CAMP_BUZZ_VHS_ROOT/env.sh" Enter
Type "clear" Enter
Show

Type "camp-buzz doctor" Enter
Sleep 1600ms
Type "clear" Enter
Sleep 300ms

Type "camp-buzz bind --channel 11111111-1111-4111-8111-111111111111 --relay ws://localhost:3000 --festival DEMO0001" Enter
Sleep 900ms
Type "clear" Enter
Sleep 300ms

Type "camp-buzz show" Enter
Sleep 1400ms
Type "clear" Enter
Sleep 300ms

Type "camp-buzz doctor" Enter
Sleep 1600ms
Type "clear" Enter
Sleep 300ms

Type "camp-buzz post -m 'Task complete: wire camp-buzz demos' --task FEST-demo --gate pass" Enter
Sleep 1200ms
Type "clear" Enter
Sleep 300ms

Type "camp-buzz hook-install" Enter
Sleep 1800ms
Type "clear" Enter
Sleep 300ms

Type "camp buzz version" Enter
Sleep 1200ms
