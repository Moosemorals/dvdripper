
"use strict";

window.log = (function () {
    const log = document.getElementById("log")
    const levels = ["debug", "info", "warn", "error"]

    function _prepend(el) {
        log.insertBefore(el, log.firstChild)
    }

    function _log(level, ...message) {
        _prepend(buildElement("p", "log log-" + level, level.toUpperCase(), ": ", ...message))
    }

    const cmds = {}
    for (let i = 0; i < levels.length; i += 1) {
        const level = levels[i]
        cmds[level] = (...m) => _log(level, ...m)
    }
    return cmds
})()

window.Backend = (function () {
    var ws
    const queue = [];

    function _getWSPath() {
        var loc = window.location, result;
        if (loc.protocol === "https:") {
            result = "wss:";
        } else {
            result = "ws:";
        }
        result += "//" + loc.host + "/ws"

        return result
    } 

    function _queue(message) {
        log.debug("Queueing: ", message)
        if (ws.readyState === 1) { // OPEN
            _send(message)
        } else {
            queue.push(message)
        }
    }

    function _send(message) {
        log.debug("Sending: ", message)
        ws.send(message)
    }

    function _onMessage(e) {
        log.debug("Received:", e.data)
        const json = JSON.parse(e.data)
        switch (json.message) {
            case "error":
                log.error("Server error", json.payload)
            case "scan":
                handleScanResult(json.payload)
                break;
        }
    }

    function _onOpen() {
        log.debug("Websocket opened. Sending ", queue.length + " from the queue")
        while (queue.length > 0) {
            _send(queue.shift())
        }
    }

    function _init() {
        ws = new WebSocket(_getWSPath())
        ws.onopen = _onOpen
        ws.onmessage = _onMessage
    }

    window.addEventListener("DOMContentLoaded", _init)

    return {
        send: _queue, 
    }
})()


function textNode(text) {
    return document.createTextNode(text)
} 

function appendChildren(el, ...children) {
    for (let i = 0; i < children.length; i += 1) {
        const arg = children[i]
        switch (typeof arg) {
            case "string":
            case "number":
            case "boolean":
                el.appendChild(textNode(arg))
                break;
            case "object":
                if (arg instanceof HTMLElement) {
                    el.appendChild(arg)
                } else {
                    el.appendChild(textNode(JSON.stringify(arg)))
                }
                break;
        }
    }

    return el
}

function buildFragment(...contents) {
    return appendChildren(document.createDocumentFragment(), ...contents)
}

function buildElement(tag, options, ...contents) {
    const el = document.createElement(tag)

    switch (typeof options) {
        case "string":
            el.setAttribute("class", options)
            break
        case "object":
            for (let key in options) {
                if (options.hasOwnProperty(key) && options[key] !== undefined) {
                    el.setAttribute(key, options[key])
                }
            }
            break;
    }

    return appendChildren(el, ...contents)
}

function empty(el) {
    while (el.firstChild) {
        el.removeChild(el.firstChild)
    }
    return el
}

function buildFilename(diskId, track) {
    return `${diskId} - Track ${track.id} (${track.length})`
}

function buildTrackRow(diskId, track) {
    return buildElement("tr", { "data-track": track.id },
        buildElement("td", undefined,
            buildElement("input", { type: "checkbox" })
        ),
        buildElement("td", undefined, track.id),
        buildElement("td", undefined, track.chapter),
        buildElement("td", undefined, track.length),
        buildElement("td", undefined,
            buildElement("input", { type: "text", value: buildFilename(diskId, track) })
        )
    )
}

function buildTrackRows(scan) {
    const rows = buildFragment()
    for (let i = 0; i < scan.tracks.length; i += 1) {
        rows.appendChild(buildTrackRow(scan.diskId, scan.tracks[i]))
    }
    return rows
}

function handleScanResult(scan) {
    empty(document.getElementById("tracklist")).appendChild(buildTrackRows(scan))

    document.getElementById("cmd-rip").disabled = false
}

function cmdSend() {
    log.info("Requesting scan")
    Backend.send("scan")
}

document.getElementById("cmd-scan").addEventListener("click", cmdSend)