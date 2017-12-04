{{define "method-js"}}

var id = {{.Gadget.Id}}
var inProgressKey = id + "-in-progress-method";

function confirm(step) {
    var msg = JSON.stringify({
        type: 'method update',
        sender: 'quimby',
        body:step
    });
    ws.send(msg);
}

function addStep(ul, text, i, step, time) {
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
    } else if (i == step && time) {
        var a2 = document.createElement("a");
        a2.setAttribute("class", "confirm");
        a2.text = time;
        li.appendChild(a2);
    }
    ul.appendChild(li);
}

function showMethod(method) {
	if (method.step > 0) {
		localStorage.setItem(inProgressKey, JSON.stringify(method));
	}
	
    var ul = document.getElementById("steps");
    while (ul.firstChild) {
        ul.removeChild(ul.firstChild);
    }

    _.each(method.steps, function(step, i) {
        addStep(ul, step, i, method.step, method.time);
    })
}

{{end}}
