{{define "delete-js"}}

var resource = {{.Resource}};

function doDelete() {
    document.getElementById('delete-form').onsubmit = function() {
        return false;
    };
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            window.location.href = "/admin.html";
        }
    }
    xhr.open("DELETE", resource, true); // true for asynchronous 
    xhr.send(null);
    return false;
}

{{end}}

