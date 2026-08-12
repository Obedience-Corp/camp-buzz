# Full-color TTY recording: camp-buzz posting through the REAL buzz CLI
# against a live local relay.
#
# Prerequisites (not recorded):
#   just relay                    # in Buzz monorepo
#   export BUZZ_BIN=…/target/release/buzz
#   export BUZZ_PRIVATE_KEY=…     # e.e. e2e tyler key or buzz-admin generate-key
#   just vhs-cli-real
#
# Outputs GIF + MP4 for README.

Output docs/demos/camp-buzz-cli-real.gif
Output docs/demos/camp-buzz-cli-real.mp4

Set Shell "bash"
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 15
Set Width 1100
Set Height 640
Set Padding 16
Set Margin 0
Set BorderRadius 0
Set Framerate 24
Set TypingSpeed 28ms
Set Theme {"name":"termcast","background":"#0d0d11","foreground":"#f8f8f2","cursor":"#F2721C","selection":"#2a2a33","black":"#16161a","red":"#ff5f57","green":"#50fa7b","yellow":"#febc2e","blue":"#57a6ff","magenta":"#bd93f9","cyan":"#8be9fd","white":"#f8f8f2","brightBlack":"#6b7280","brightRed":"#ff6e67","brightGreen":"#69ff94","brightYellow":"#ffd23f","brightBlue":"#7ab8ff","brightMagenta":"#d0a9ff","brightCyan":"#a4f0ff","brightWhite":"#ffffff"}

Hide
Type "source /tmp/camp-buzz-cli-real-demo/env.sh" Enter
Type "clear" Enter
Show

Type "camp-buzz doctor" Enter
Sleep 1600ms
Type "clear" Enter
Sleep 300ms

Type "camp-buzz post -m 'Hello from camp-buzz — real relay + UI next' --task FEST-ui --gate pass" Enter
Sleep 1800ms
Type "clear" Enter
Sleep 300ms

Type "buzz messages get --channel $CAMP_BUZZ_DEMO_CHANNEL --limit 2" Enter
Sleep 2800ms
