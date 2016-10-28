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
        document.getElementById(id).textContent = msg.value.value;
    }
});

function getValue(v) {
    if (isNumeric(v)) {
        return num.toFixed(v);
    }
    return v
}

function isNumeric(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

{{end}}
