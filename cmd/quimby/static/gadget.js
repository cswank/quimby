{{define "gadget-js"}}

var ws;

function getWebsocket() {
    var url = "{{ .Websocket }}";
    return new WebSocket(url);
}

function doConnect(callback) {
    if(ws != undefined) {
        ws.close();
        ws = null;
    }
    ws = getWebsocket();
    ws.onopen = function() {};
    ws.onerror = function(data) {console.log("error", data)};
    ws.onmessage = function(message) {
        message = JSON.parse(message.data);
        callback(message);
    };
}

doConnect(function(msg) {
    if (msg.type == "update") {
        var id = msg.location + "-" + msg.name;
        document.getElementById(id).textContent = getValue(msg.value.value);
    }
});

function getValue(v) {
    if (isNumeric(v)) {
        return v.toFixed(1);
    }
    return v
}

function isNumeric(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

function sendCommand(id, info) {
    var cmd;
    if (document.getElementById(id).textContent == "true") {
        cmd = info.off[0];
    } else {
        cmd = info.on[0];
    }
    var msg = JSON.stringify({
        sender: "quimby",
        type: "command",
        body: cmd,
    });
    ws.send(msg);
}
{{end}}

