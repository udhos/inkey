# inkey
non-blocking keyboard input

# Install

    git clone https://github.com/udhos/inkey
    cd inkey
    go install ./inkey-run

# Running 'inkey-run' test app

    stty cbreak min 1 ;# disable input buffering

    inkey-run

    stty sane ;# restore terminal defaults
