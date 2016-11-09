{{define "edit-gadget-js"}}

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

function setAction(name) {
    var action = actions[name];
    if (name != "submit") {
        document.getElementById("name").removeAttribute("required");
        document.getElementById("host").removeAttribute("required");
    }
    document.getElementById("gadget-form").setAttribute("action", action.uri);
    document.getElementById("gadget-form").setAttribute("method", action.method);
    document.getElementById("gadget-form").submit();
    return false;
}
{{end}}

