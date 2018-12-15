
function getWSPath() {
    var loc = window.location, result;
    if (loc.protocol === "https:") {
        result = "wss:";
    } else {
        result = "ws:";
    }
    result += "//" + loc.host + "/ws"

    return result
}

function textNode(text) {
    return document.createTextNode(text)
}

function buildElement(tag, options, ...contents) {
    const el = document.createElement(tag)

    switch (typeof options) {
        case "string":
            el.setAttribute("class", options)
            break
        case "object":
            for (let key in object) {
                if (object.hasOwnProperty(key) && object[key] !== undefined) {
                    el.setAttribute(key, object[key])
                }
            }
            break;
    }

    for (let i = 0; i < contents.length; i += 1) {
        const arg = contents[i]
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

function empty(el) {
    while (el.firstChild) {
        el.removeChild(el.firstChild)
    }
    return el
}

function showScanTable(scan) {
    const tbody = buildElement("tbody")

    for (let i = 0; i < scan.Titles.length; i += 1) {
        const track = scan.Titles[i];

        tbody.appendChild(
            buildElement("tr", undefined,
                buildElement("td", undefined, track.Title),
                buildElement("td", undefined, track.Length),
                buildElement("td", undefined, track.Chapters)
            )
        )
    }

    empty(document.getElementById("out")).appendChild(
        buildElement("table", undefined,
            buildElement("thead", undefined,
                buildElement("tr", undefined,
                    buildElement("th", undefined, "Track"),
                    buildElement("th", undefined, "Length"),
                    buildElement("th", undefined, "Chapters")
                )
            ),
            tbody
        )
    )
}


function init() {
    const ws = new WebSocket(getWSPath())

    ws.onmessage = e => {
        json = JSON.parse(e.data)
        switch (json.Message) {
            case "scan":
                showScanTable(json.Payload)
                break;
        }
    }

    ws.onopen = () => {
        ws.send(JSON.stringify("Scan"))
    }
}

window.addEventListener("DOMContentLoaded", init)