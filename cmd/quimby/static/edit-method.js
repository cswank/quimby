{{define "edit-method-js"}}

var id = {{.Gadget.Id}}
var methodUrl = "/api/gadgets/" + id + "/method";
var key = id + "-methods";

localStorage.setItem(key, JSON.stringify({test: {steps: ["turn on back yard thing", "wait for 5 seconds", "turn off back yard thing"]}}));

var methods = getStoredMethods();
showStoredMethods();

function runMethod() {
    var select = document.getElementById("stored-methods");
    var val = select.options[select.selectedIndex].value;
    var method = methods[val];
    postMethod({method: method.steps});
    return true;
}

function postMethod(method) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", methodUrl, true);
    xhr.setRequestHeader("Content-type", "application/json");
    xhr.send(JSON.stringify(method));
}

function showMethod() {
    var select = document.getElementById("stored-methods");
    var val = select.options[select.selectedIndex].value;
    document.getElementById("method").value = methods[val].steps.join("\n");
}

function showStoredMethods() {
    var select = document.getElementById("stored-methods");
    console.log("methods", methods);
    _.each(methods, function(val, key) {
        var opt = document.createElement("option");
        opt.setAttribute("value", key);
        opt.text = key;
        console.log(opt);
        select.appendChild(opt);
    })
}

function getStoredMethods() {
    var m = JSON.parse(localStorage.getItem(key));
    console.log(m);
    if (m == null) {
        m = [];
    }
    return m
}

{{end}}
