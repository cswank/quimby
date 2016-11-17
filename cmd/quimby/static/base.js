{{define "base-js"}}

var ws = new WebSocket("{{.Websocket}}");

window.onbeforeunload = function() {
    ws.onclose = function () {};
    ws.close();
};

ws.onerror = function(data) {console.log("error", data)};

function updateIO(msg) {
    var id = msg.location + "-" + msg.name;
    if (msg.info.direction == "input") {
        document.getElementById(id).textContent = getValue(msg.value.value);
    } else if (msg.info.direction == "output") {
        document.getElementById(id).checked = msg.value.value;
    }
}

function getValue(v) {
    if (isNumeric(v)) {
        return v.toFixed(1);
    }
    return v
}

function isNumeric(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

function doSendComamnd(cmd) {
    var msg = JSON.stringify({
        sender: "quimby",
        type: "command",
        body: cmd,
    });
    ws.send(msg);
}

function waitForSocketConnection(ws, callback) {
    setTimeout(
        function () {
            if (ws.readyState === 1) {
                if(callback != null){
                    callback();
                }
                return;
            } else {
                waitForSocketConnection(ws, callback);
            }
        }, 50);
}

{{end}}

