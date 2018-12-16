
"use strict";

window.log = (function () {
    const log = document.getElementById("log")
    const levels = ["debug", "info", "warn", "error"]
    let currentLevel = "info"

    function _prepend(el) {
        log.insertBefore(el, log.firstChild)
    }

    function _log(level, ...message) {
        const el =buildElement("p", "log log-" + level, level.toUpperCase(), ": ", ...message) 

        if (_levelToInt(level) < _levelToInt(currentLevel)) {
            el.classList.add("hidden")
        }
        _prepend(el)
    }

    function _levelToInt(level) {
        return levels.indexOf(level)
    }

    function _setLevel(target) {
        currentLevel = target
        let show = true
        for (let i = levels.length - 1; i >= 0; i -= 1) {
            const level = levels[i]
            if (show) {
                $$(".log-" + level).forEach(x => x.classList.remove("hidden"))
            } else { 
                $$(".log-" + level).forEach(x => x.classList.add("hidden"))
            }
            if (level === target) {
                show = false
            } 
        }
    }

    document.getElementById("log-level").addEventListener("change", e => _setLevel(e.target.value.toLowerCase()))

    const cmds = {
        setLevel: _setLevel
    }
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

    function _queue(cmd, payload) {
        const packet = {
            cmd: cmd,
            payload: payload
        }
        log.debug("Queueing: ", packet)


        if (ws.readyState === 1) { // OPEN
            _send(packet)
        } else {
            queue.push(packet)
        }
    }

    function _send(packet) {
        log.debug("Sending: ", packet)
        ws.send(JSON.stringify(packet))
    }

    function _onMessage(e) {
        log.debug("Received:", e.data)
        const json = JSON.parse(e.data)
        switch (json.cmd) {
            case "error":
                log.error("Server error", json.payload)
                break;
            case "rip-started":
                handleRipStarted(json.payload)
                break;
            case "rip-progress":
                handleRipProgress(json.payload)
                break;
            case "rip-completed":
                handleRipCompleted(json.payload)
                break;
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

/**
 * $$ returns all elements that match selector
 * @param {string} selector 
 * @returns {[HTMLElement]}
 */
function $$(selector) { return Array.from(document.querySelectorAll(selector)) }

/**
 * Removes any child content from el. Returns el for chaining
 * @param {HTMLElement} el 
 * @returns {HTMLElement}
 */
function empty(el) {
    while (el.firstChild) {
        el.removeChild(el.firstChild)
    }
    return el
}
/**
 * Wraps text in a DOM text node 
 * @param {string|number|boolean} text 
 * @returns {HTMLText}
 */
function textNode(text) {
    return document.createTextNode(text)
}

/**
 * Appends anyhting to el, converting to DOM objects as needed. Returns el for chaining
 * @param {HTMLElement} el 
 * @param  {...any} children 
 * @returns {HTMLElement}
 */
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

/**
 * Creates a new DocumentFragment. Appends contents to the newly formed fragment.
 * @param  {...any} contents 
 * @readonly {HTMLDocumentFragment}
 */
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

function buildFilename(diskId, track) {
    return `${diskId} - Track ${track.id}.mpeg`
}

function buildTrackRow(diskId, track) {
    return buildElement("tr", undefined,
        buildElement("td", undefined,
            buildElement("input", { class: "rip-check", type: "checkbox", id: "chk-" + track.id })
        ),
        buildElement("td", undefined, track.id),
        buildElement("td", undefined, track.chapter),
        buildElement("td", undefined, track.length),
        buildElement("td", undefined,
            buildElement("input", { type: "text", value: buildFilename(diskId, track), class: "track-filename", id: "ipt-" + track.id })
        ),
        buildElement("td", undefined,
            buildElement("progress", { max: 100, value: 0, id: "progress-" + track.id, class: "download-class", title: "0%" }, "0%")
        ),
        buildElement("td", { id: "download-" + track.id })
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
    log.info("Scan complete")
    empty(document.getElementById("tracklist")).appendChild(buildTrackRows(scan))

    document.getElementById("cmd-rip").disabled = false
}

function cmdSend() {
    log.info("Requesting scan")
    Backend.send("scan")
}

function updateProgress(progressBar, percent) {
    const percentString = percent + "%"

    progressBar.value = percent
    empty(progressBar).appendChild(textNode(percentString))
    progressBar.title = percentString
}

function handleRipStarted(payload) {
    log.info("Started ripping track ", payload.track)
    updateProgress($$("#progress-" + payload.track)[0], 0)
}

function handleRipProgress(payload) {
    updateProgress($$("#progress-" + payload.track)[0], payload.percent)
}

function handleRipCompleted(payload) {
    log.info("Completed ripping track ", payload.track)
    updateProgress($$("#progress-" + payload.track)[0], 100)

    const chk = document.getElementById("chk-" + payload.track)
    chk.checked = false
    chk.disabled = true

    document
        .getElementById("download-" + payload.track)
        .appendChild(
            buildElement("a", {
                href: "rips/" + payload.filename
            }, 
            "Download" 
            )
        )
}

function cmdRip() {
    const tracks = $$("[type=checkbox]:checked").map(x => {
        const id = x.id.substring("chk-".length)
        return {
            track: parseInt(id, 10),
            filename: document.getElementById("ipt-" + id).value
        }
    })

    log.debug("Want to rip ", tracks)
    Backend.send("rip", tracks)
}

document.getElementById("cmd-scan").addEventListener("click", cmdSend)
document.getElementById("cmd-rip").addEventListener("click", cmdRip)