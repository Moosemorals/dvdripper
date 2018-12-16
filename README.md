# DVDRipper - A wrapper around a copule of tools to make ripping Dr Who easier

## Dependencies

    sudo apt install lsdvd mplayer

Install libdvdcss from [Videolan](http://www.videolan.org/developers/libdvdcss.html)

    sudo apt install libdvdcss2

## Status

So far so good. I've got the front end displaying the scan from the DVD, next is actually ripping.

I'm going to get the frontend to do most of the sequencing, so it will be sending a series of
"rip" commands, something like

    cmd: "rip",
    payload: {track: 3, filename: "Track 3"}

It will get back a series of updates:

    cmd: "rip-start"
    payload: {track: 3, filename: "Track 3"}    // Not sure if I need the filenmae

    cmd: "rip-progress"
    payload: {track: 3, bytes: 12345, percent: 2.4}

    cmd: "rip-complete"
    payload: {track: 3, url: "/rips/Track 3"}

    cmd: "download-complete"
    payload: {track: 3}