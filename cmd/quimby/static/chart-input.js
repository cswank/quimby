{{define "chart-input-js"}}

var key = {{.Key}};
var back = {{.Back}};

function save() {
    var checked = document.getElementById("binary").checked ? "true" : "false";
    localStorage.setItem(key, checked);
    window.location.href = back;
}

var checked = (localStorage.getItem(key) == "true") ? true : false;
document.getElementById("binary").checked = checked;

{{end}}
