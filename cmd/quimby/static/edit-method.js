{{define "edit-method-js"}}

var id = {{.Gadget.Id}}
var methodUrl = "/api/gadgets/" + id + "/method";
var brewMethodUrl = "/api/beer/"
var key = id + "-methods";
var inProgressKey = id + "-in-progress-method";

var methods = getStoredMethods();
showStoredMethods();

function runMethod() {
	var title = document.getElementById("title").value;
	var method;
	if (title == "last method") {
		method = JSON.parse(localStorage.getItem(inProgressKey));
		localStorage.setItem(inProgressKey, null);
		postMethod(method);
	} else {
		method = doSaveMethod();
		postMethod({steps: method.steps});
	}
    return true;
}

function deleteMethod() {
    var select = document.getElementById("stored-methods");
    var val = select.options[select.selectedIndex].value;
    delete methods[val];
    localStorage.setItem(key, JSON.stringify(methods));
    document.getElementById("title").value = "";
    document.getElementById("method").value = "";

    if (select.children.length > 0) {
        while (select.firstChild) {
            select.removeChild(select.firstChild);
        }
    }
    showStoredMethods();
    document.getElementById('method-form').onsubmit = function() {
        return false;
    };
    return false;
}

function getBrewMethod() {
    var name = document.getElementById("brew-name").value;
    var temp = document.getElementById("grain-temperature").value;
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            var data = JSON.parse(xhr.responseText);
            document.getElementById("method").value = data.join("\n");
            document.getElementById("title").value = name;
        }
    }
    var url = brewMethodUrl + name + "?grain_temperature=" + temp;
    xhr.open("GET", url, true); // true for asynchronous 
    xhr.send(null);
    return false;
}

function postMethod(method) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", methodUrl, false);
    xhr.setRequestHeader("Content-type", "application/json");
    xhr.send(JSON.stringify(method));
}

function saveMethod() {
    doSaveMethod();
    document.getElementById('method-form').onsubmit = function() {
        return false;
    };
    return false;
}

function doSaveMethod() {
    var title = document.getElementById("title").value;
    var method = {steps: document.getElementById("method").value.split("\n")};
    methods[title] = method;
    localStorage.setItem(key, JSON.stringify(methods));
    return method;
}

function showMethod() {
    var select = document.getElementById("stored-methods");
    var val = select.options[select.selectedIndex].value;
    if (val == "") {
        document.getElementById("title").value = "";
        document.getElementById("method").value = "";
    } else {
        document.getElementById("title").value = val;
        document.getElementById("method").value = methods[val].steps.join("\n");
    }
}

function showInProgressMethod() {
	var method = JSON.parse(localStorage.getItem(inProgressKey));
	if (method != null) {
		document.getElementById("title").value = "last method";
		document.getElementById("method").value = method.steps.join("\n");
	}
}

function showStoredMethods() {
    var select = document.getElementById("stored-methods");
    var opt = document.createElement("option");
    select.appendChild(opt);
    _.each(methods, function(val, key) {
        opt = document.createElement("option");
        opt.setAttribute("value", key);
        opt.text = key;
        select.appendChild(opt);
    })
}

function getStoredMethods() {
    var m = JSON.parse(localStorage.getItem(key));
    if (m == null) {
        m = {};
    }
    return m
}

{{end}}
