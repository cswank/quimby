{{define "edit-user-js"}}

var actions = {
    {{$end := .End}}
    {{range $index, $val := .Actions}}
    {{if eq $index $end}}
    "{{$val.Name}}": {"uri": "{{$val.URI}}", "method": "{{$val.Method}}"}
    {{else}}
    "{{$val.Name}}": {"uri": "{{$val.URI}}", "method": "{{$val.Method}}"},
    {{end}}
    {{end}}
};

function setAction(name, isNew) {
    var action = actions[name];
    if (name != "submit" && isNew == "true") {
        document.getElementById("username").removeAttribute("required");
        document.getElementById("password").removeAttribute("required");
        document.getElementById("password-confirm").removeAttribute("required");
    }
    document.getElementById("user-form").setAttribute("action", action.uri);
    document.getElementById("user-form").setAttribute("method",action.method);
    document.getElementById("user-form").submit();
}
{{end}}

