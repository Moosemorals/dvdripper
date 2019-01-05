"use strict";

window.log = (function () {
    const log = $("#log")
    const levels = ["debug", "info", "warn", "error"]
    let currentLevel = "info"

    function _prepend(el) {
        log.insertBefore(el, log.firstChild)
    }

    function _log(level, ...message) {
        const el = buildElement("p", "log log-" + level, level.toUpperCase(), ": ", ...message)

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

    $("#log-level").addEventListener("change", e => _setLevel(e.target.value.toLowerCase()))

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
        var loc = window.location,
            result;
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
            case "freespace":
                handleFreespace(json.payload)
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
function $$(selector) {
    return Array.from(document.querySelectorAll(selector))
}

function $(selector) {
    return document.querySelector(selector)
}

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

function pad2(num) {
    if (num < 10) {
        return "0" + num
    } else {
        return "" + num
    }
}

function buildFilename(diskId, track) {
    return `${diskId} - Track ${pad2(track.id)}.mpg`
}

function buildTrackRow(diskId, track) {
    return buildElement("tr", undefined,
        buildElement("td", undefined,
            buildElement("input", {
                class: "rip-check",
                type: "checkbox",
                id: "chk-" + track.id
            })
        ),
        buildElement("td", "text-center", track.id),
        buildElement("td", "text-center", track.chapters),
        buildElement("td", undefined, track.length),
        buildElement("td", undefined,
            buildElement("input", {
                type: "text",
                value: buildFilename(diskId, track),
                class: "track-filename",
                id: "ipt-" + track.id
            })
        ),
        buildElement("td", undefined,
            buildElement("progress", {
                max: 100,
                value: 0,
                id: "progress-" + track.id,
                class: "download-class",
                title: "0%"
            }, "0%")
        ),
        buildElement("td", {
            class: "text-center",
            id: "status-" + track.id
        }, "-"),
    )
}

function buildTrackRows(scan) {
    const rows = buildFragment()
    for (let i = 0; i < scan.tracks.length; i += 1) {
        rows.appendChild(buildTrackRow(scan.diskId, scan.tracks[i]))
    }
    return rows
}

function buildErrorRow(message) {
    return buildElement("tr", undefined, buildElement("td", {colspan:5}, message))
}

function handleScanResult(scan) {
    setContent($("#cmd-scan"), "Scan")
    log.info("Scan complete")
    if (scan.tracks !== null) {
        $("#track-all").value = `${scan.diskId} - Track`
        empty($("#tracklist")).appendChild(buildTrackRows(scan))
        $("#cmd-rip").disabled = false
        $("#cmd-scan").disabled = false
    } else {
        empty($("#tracklist")).appendChild(buildErrorRow("No tracks found"))
    }
     $("#output").classList.remove("hidden")
}

function setContent(el, ...content) {
    appendChildren(empty(el), ...content)
}

function cmdScan() {
    log.info("Requesting scan")
    setContent(this, "Scanning")
    this.disabled = true
    Backend.send("scan")
}

function updateProgress(payload, percent) {
    if (percent === undefined) {
        percent = payload.percent
    }
    let percentString
    const progressBar = $("#progress-" + payload.track)
    if (percent !== -1) {
        percentString = percent + "%"
        progressBar.value = percent
    } else {
        percentString = "-%"
        progressBar.removeAttribute("value")
    }

    progressBar.title = percentString
    setContent($("#status-" + payload.track), "Ripping: ", percentString)
    setContent(progressBar, percentString)
    
}

function handleRipStarted(payload) {
    setContent($("#status-" + payload.track), "Starting rip")
    log.info("Started ripping track ", payload.track)
    updateProgress(payload, 0)
}

function handleRipProgress(payload) {
    updateProgress(payload)
}

function handleFreespace(payload) {
    const fs = $("#fs")
    fs.min = 0
    fs.max = payload.total
    fs.value = payload.total - payload.free
}

function notify(payload) {
    if (Notification.permission === "granted") {
        new Notification("Track " + payload.track + " has finished ripping and is ready to download")
    } else if (Notification.permission !== "denied") {
        Notification.requestPermission().then(permission => {
            if (permission === "granted") {
                notify(payload)
            }
        })
    }
}

function handleRipCompleted(payload) {
    log.info("Completed ripping track ", payload.track)
    updateProgress(payload, 100)

    const chk = $("#chk-" + payload.track)
    chk.checked = false
    chk.disabled = true

    setContent($("#status-" + payload.track),
        buildElement("a", {
                href: "rips/" + payload.filename
            },
            "Download"
        )
    )

    notify(payload)
}

function cmdRip() {
    this.disabled = true
    const tracks = $$(".rip-check:checked").map(x => {
        const id = x.id.substring("chk-".length)
        setContent($("#status-" + id), "Queued")
        return {
            track: parseInt(id, 10),
            filename: $("#ipt-" + id).value
        }
    })

    log.debug("Want to rip ", tracks)
    Backend.send("rip", tracks)
}

function toggleAll(e) {
    $("#chk-all").indeterminate = false
    $$(".rip-check").forEach(x => x.checked = this.checked)
}

function ripCheck(e) {
    const $all = $("#chk-all")
    const state = $all.checked
    let indi = false
    $$(".rip-check").forEach(x => {
        if (x.checked !== state) {
            indi = true
        }
    });
    $all.indeterminate = indi

    const id = this.id.substring("chk-".length)
    setContent($("#status-" + id), this.checked ? "Selected" : "")
}

function cmdStop() {
    Backend.send("interrupt")
}

function cmdEject() {
    Backend.send("eject")
}

function cmdTidy() {
    Backend.send("tidy")
}

const clickHandlers = {
    "#chk-all": toggleAll,
    ".rip-check": ripCheck,
    "#cmd-scan": cmdScan,
    "#cmd-rip": cmdRip,
    "#cmd-eject": cmdEject,
    "#cmd-tidy": cmdTidy
}

document.addEventListener("click", e => {
    for (let key in clickHandlers) {
        const target = e.target.closest(key);

        if (target !== null) {
            clickHandlers[key].call(target, e)
            return;
        }
    }
})

$("#track-all").addEventListener("input", e=> {
    const name = e.target.value
    $$("#tracklist .track-filename").forEach((ipt, i) => {
        ipt.value = `${name} ${pad2(i)}.mpg`
    })
})
