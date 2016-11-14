{{define "gadget-js"}}

var id = {{.Gadget.Id}}

function updateIO(msg) {
    var id = msg.location + "-" + msg.name;
    if (msg.info.direction == "input") {
        document.getElementById(id).textContent = getValue(msg.value.value);
    } else if (msg.info.direction == "output") {
        document.getElementById(id).checked = msg.value.value;
    }
}

doConnect(function(msg) {
    if (msg.type == "update") {
        updateIO(msg);
    } else if (msg.type == "method update") {
        showMethod(msg);
    }
});

function sendCommand(id, info) {
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

{{end}}
