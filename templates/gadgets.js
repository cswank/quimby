{{define "gadgets.js"}}

var ready = false;
var timeoutID;
var ws = new WebSocket("{{.Websocket}}");
var holdTime = 1000;
var commands = {{ command .Gadget.Status }};

window.onbeforeunload = function() {
    ws.onclose = function () {};
    ws.close();
};

ws.onerror = function(data) {console.log("error", data);};

function updateIO(msg) {
    var id = msg.location + "-" + msg.name;
    if (msg.info.direction == "input") {
        document.getElementById(id).textContent = getValue(msg.value.value);
    } else if (msg.info.direction == "output") {
        document.getElementById(id).checked = msg.value.value;
        console.log("msg", msg.value.io.gpio);
        if (msg.value.io.gpio) {
            document.getElementById(id + "-slider").classList.add("io-on");
        } else {
            document.getElementById(id + "-slider").classList.remove("io-on");
        }
        document.getElementById(id).checked = msg.value.value;
    }
}

function getValue(v) {
    if (isNumeric(v)) {
        return v.toFixed(2);
    }
    return v
}

function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function doSendCommand(cmd) {
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

var id =  3 ;
var ready = false;

function sendCommand(id) {
    var info = commands[id].info;
    if (!ready) {
        showNotReady(id);
    }
    var cmd;
    if (document.getElementById(id).checked) {
        cmd = info.on[0];
    } else {
        cmd = info.off[0];
    }
    
    doSendCommand(cmd);    
}

waitForSocketConnection(ws, function() {
    ready = true;
    doSendCommand("update");

    var devices = document.getElementsByClassName("device");
    _.each(devices, function(dev) {
        dev.addEventListener('mousedown', function(event) { 
            timeoutId = setTimeout(function() {
                //dev.
                showCommand(dev);
            }, holdTime);
            dev.addEventListener('mouseup', function(event) {
                clearTimeout(timeoutId);
            });
        });
    });
});

ws.onmessage = function(message) {
    var msg = JSON.parse(message.data);
    if ((msg.type == "update" && msg.sender == "method runner") || msg.type == "method update") {
        showMethod(msg.method);
    } else if (msg.type == "update") {
        updateIO(msg);
    }
};


function showCommand(label) {
    var dev = label.getElementsByTagName("input")[0];
    var msg = commands[dev.id];
    
    var state;
    if (dev.checked) {
        state = "off";
    } else {
        state = "on";
    }

    var cmd = prompt("send a command", msg.info[state]);
    doSendCommand(cmd);
}

{{end}}
