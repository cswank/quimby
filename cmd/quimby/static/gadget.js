{{define "gadget-js"}}

var id = {{.Gadget.Id}}

doConnect(function(msg) {
    if (msg.type == "update") {
        var id = msg.location + "-" + msg.name;
        if (msg.info.direction == "input") {
            document.getElementById(id).textContent = getValue(msg.value.value);
        } else if (msg.info.direction == "output") {
            document.getElementById(id).checked = msg.value.value;
        }
    }
});

function sendCommand(id, info) {
    var cmd;
    if (document.getElementById(id).checked) {
        cmd = info.on[0];
    } else {
        cmd = info.off[0];
    }
    doSendComamnd(msg);
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

{{end}}
