# DVDRipper - A wrapper around a copule of tools to make ripping Dr Who easier

## Dependencies

    sudo apt install lsdvd mplayer

Install libdvdcss from [Videolan](http://www.videolan.org/developers/libdvdcss.html)

    sudo apt install libdvdcss2

## Status

I've got wrappers round lsdvd and mplayer. Next I need to sort out an interface.

Of course, it's going to be http...

open a websocket

send a scan request
get scan response

send a rip request
get progress
