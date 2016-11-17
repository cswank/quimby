{{define "gadget-js"}}

var id = {{.Gadget.Id}}
var ready = false;

function sendCommand(id, info) {
    if (!ready) {
        showNotReady(id);
    }
    var cmd;
    if (document.getElementById(id).checked) {
        cmd = info.on[0];
    } else {
        cmd = info.off[0];
    }
    doSendComamnd(cmd);
}

function showChart(location, name) {
    var url = "/api/gadgets/" + id + "/sources/" + location + "%20" + name;
    var href = window.location.href + "/chart.html?source=" + url;
    window.location.href = href;
}

function showChartSetup() {
    window.location.href = window.location.href + "/chart-setup.html";
    return false;
}

function showNotReady(id) {
    document.getElementById(id).checked = !document.getElementById(id).checked;
    document.getElementById("not-ready").text = "not connected";
    setTimeout(
        function () {
            document.getElementById("not-ready").text = "";
        }, 1000);
}

waitForSocketConnection(ws, function() {
    ready = true;
    doSendComamnd("update");
});

ws.onmessage = function(message) {
    var msg = JSON.parse(message.data);
    if ((msg.type == "update" && msg.sender == "method runner") || msg.type == "method update") {
        showMethod(msg.method);
    } else if (msg.type == "update") {
        updateIO(msg);
    }
};

{{end}}
