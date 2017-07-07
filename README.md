# antgo
ANT+ Library in GO

## Example usage:
- Start a websocket server with: `./antserver`
- Start client with capture file: `./antdump -driver=file -infile=123.cap -ws="ws://localhost:8080"`
- Or with usb dongle: `./antdump -driver=usb -pid=0x1008 -ws="ws://localhost:8080"`
- Or if you want to dump to a file: `./antdump -outfile=mydump.cap`
- You can use the `-silent` switch to disable ANT+ data output on the terminal
