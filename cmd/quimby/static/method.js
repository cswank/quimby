{{define "method-js"}}

function confirm(step) {
    var msg = {
        type: 'method update',
        sender: 'client',
        body:step
    };
    ws.send(msg);
}

function addStep(text, i, step) {
    var ul = document.getElementById("steps");
    var li = document.createElement("li");
    li.setAttribute("class", "step");
    var a = document.createElement("a");
    a.text = text;
    if (i < step) {
        a.setAttribute("class", "complete");
    }
    li.appendChild(a);
    if (i == step && text.indexOf("wait for user") == 0) {
        var a2 = document.createElement("a");
        a2.text = "confirm";
        a2.setAttribute("class", "confirm");
        a2.setAttribute("onClick", "confirm('" + text + "')");
        li.appendChild(a2);
    }
    ul.appendChild(li);
}

function showMethod(method) {
    _.each(method.steps, function(step, i) {
        addStep(step, i, method.step);
    })
}

var method = {
    steps: [
        "turn on file",
        "wait for user to feel good",
        "turn off file"
    ],
    step: 1,
}

showMethod(method);

{{end}}
