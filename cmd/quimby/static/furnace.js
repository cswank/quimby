{{define "furnace-js"}}

function setState(val) {
    if (val.command == "heat home") {
        document.getElementById("furnace-heat").checked = true;
        if (val.io.heat) {
            document.getElementById("heat-label").setAttribute("class", "on");
        } else {
            document.getElementById("heat-label").setAttribute("class", "off");
        }
        document.getElementById("cool-label").removeAttribute("class");
    } else if (val.command == "cool home") {
        document.getElementById("furnace-cool").checked = true;
        if (val.io.heat) {
            document.getElementById("cool-label").setAttribute("class", "on");
        } else {
            document.getElementById("cool-label").setAttribute("class", "off");
        }
        document.getElementById("heat-label").removeAttribute("class");
    } else {
        document.getElementById("furnace-off").checked = true;
        document.getElementById("cool-label").removeAttribute("class");
        document.getElementById("heat-label").removeAttribute("class");
    }
}

function updateState() {
    var cmd = getCommand();
    doSendComamnd(cmd);
}

function setPointChange() {
    updateState();
}

function updateSetPointDisplay() {
    document.getElementById("set-point-display").textContent = document.getElementById("set-point").value;
}

function getCommand() {
    var sp = document.getElementById("set-point").value;
    var state = document.querySelector('input[name="state"]:checked').value;
    var cmd = "turn off furnace";
    if (state == "heat") {
        cmd = "heat home to " + sp + " F";
    } else if (state == "cool") {
        cmd = "cool home to " + sp + " F";
    }
    return cmd;
}

setState({{.Furnace.Value}});

doConnect(function(msg) {
    if (msg.type == "update") {
        var id = msg.location + "-" + msg.name;
        if (id == "home-temperature") {
            document.getElementById(id).textContent = getValue(msg.value.value);
        } else if (id == "home-furnace") {
            setState(msg.value);
        }
    }
});

{{end}}
