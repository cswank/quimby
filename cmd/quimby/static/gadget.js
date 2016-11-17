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
