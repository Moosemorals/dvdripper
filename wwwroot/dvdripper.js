
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

const ws = new WebSocket(getWSPath())

ws.onmessage = e => {
    console.log(e.data)
}

ws.onopen = () => { 
    ws.send(JSON.stringify("Scan"))
}