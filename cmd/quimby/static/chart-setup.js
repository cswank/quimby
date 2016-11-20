{{define "chart-setup-js"}}

function checkBinary() {
    var sources = document.querySelectorAll('input[name=source]:checked');
    sources.forEach(function(item) {
        var key = item.getAttribute("key");
        if (localStorage.getItem(key) == "true") {
            var val = item.value + "?binary=true";
            item.value = val;
        }
    })
}

{{end}}
