{{define "gadget-js"}}

doConnect(function(msg) {
    if (msg.type == "update") {
        var id = msg.location + "-" + msg.name;
        document.getElementById(id).textContent = getValue(msg.value.value);
        if (msg.info.direction == "output") {
            var c = "off";
            if (msg.value.value == true) {
                color = "on";
            }
            document.getElementById(id).setAttribute("class", c);
        }
    }
});

{{end}}
