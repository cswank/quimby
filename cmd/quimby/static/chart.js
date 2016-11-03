{{define "chart-js"}}

var spans = {
    hour: 1,
    day: 24,
    week: 7 * 24,
    month: 7 * 24 * 30
};

var summarize = {{.Summarize}};
var spanLabels = ["hour", "day", "week", "month"];
var selectedSpan = spans[{{.Span}}];
var sources = {{.Sources}};

function getEnd() {
    return encodeURIComponent(moment().utc().format());
}

function getStart() {
    return encodeURIComponent(moment().utc().subtract(selectedSpan, "hours").format());
}

function httpGetAsync(theUrl, callback) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            var data = JSON.parse(xhr.responseText);
            callback(data);
        }
    }
    xhr.open("GET", theUrl, true); // true for asynchronous 
    xhr.send(null);
}

function getDate() {
    return function(d){
        return d3.time.format('%x %X')(new Date(d));
    }
}

function getStats(urls, span, summarize, callback) {
    var end = moment().utc().format();
    var start = moment().utc().subtract(span, "hours").format();
    var params = {end: end, start: start, summarize: summarize};
    var vals = [];
    _.each(urls, function(url) {
        console.log(url);
        var qStr = Object.keys(params).map(function(k) {
            return encodeURIComponent(k) + '=' + encodeURIComponent(params[k])
        }).join('&');
        url = url + "?" + qStr;
        httpGetAsync(url, function(data) {
            vals.push({
                key: data.name,
                values: _.map(data.data, function (value) {
                    return {x:new Date(value.x), y:value.y};
                })
            });
            if (vals.length == urls.length) {
                callback(vals);
            }
        });
    });
}

function getRange(datum) {
    var min = 100000.0
    var max = 0.0
    _.each(datum, function(data) {
        _.each(data.values, function(value) {
            if (value.y > max) {
                max = value.y
            }
            if (value.y < min) {
                min = value.y
            }
        });
    });
    if (min > 0.0) {
        min = 0.0;
    }
    return [min, max];
}

getStats(sources, selectedSpan, summarize, function(data) {
    nv.addGraph(function() {
        var chart = nv.models.lineChart()
            .margin({left: 100})
            .useInteractiveGuideline(true)
            .showLegend(true)
            .showYAxis(true)
            .showXAxis(true)
        ;

        chart.xAxis
            .ticks(3)
            .tickFormat(getDate());

        chart.yAxis
            .tickFormat(d3.format('.02f'));

        chart.forceY(getRange(data));

        d3.select('#chart svg')
            .datum(data)
            .transition().duration(500)
            .call(chart);

        nv.utils.windowResize(function() { chart.update() });
        return chart;
    });
});

{{end}}
