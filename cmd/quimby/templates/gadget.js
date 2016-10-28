{{define "gadget-js"}}

var ws;

function getWebsocket() {
    var url = "{{ .Websocket }}";
    return new WebSocket(url);
}


function doConnect(callback) {
    console.log("connecting");
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
    console.log("ws:", msg);
    if (msg.type == "update") {
        var id = msg.location + "-" + msg.name;
        console.log(id);
        document.getElementById(id).textContent = msg.value.value;
    }
});

{{end}}

