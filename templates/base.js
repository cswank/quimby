{{define "base.js"}}

/* var ready = false;
 * var ws = new WebSocket("{{.Websocket}}");
 * 
 * window.onbeforeunload = function() {
 *     ws.onclose = function () {};
 *     ws.close();
 * };
 * 
 * ws.onerror = function(data) {console.log("error", data)};
 * 
 * function updateIO(msg) {
 *     var id = msg.location + "-" + msg.name;
 *     if (msg.info.direction == "input") {
 *         document.getElementById(id).textContent = getValue(msg.value.value);
 *     } else if (msg.info.direction == "output") {
 *         document.getElementById(id).checked = msg.value.value;
 *     }
 * }
 * 
 * function getValue(v) {
 *     if (isNumeric(v)) {
 *         return v.toFixed(2);
 *     }
 *     return v
 * }
 * 
 * function isNumeric(n) {
 *   return !isNaN(parseFloat(n)) && isFinite(n);
 * }
 * 
 * function doSendComamnd(cmd) {
 *     var msg = JSON.stringify({
 *         sender: "quimby",
 *         type: "command",
 *         body: cmd,
 *     });
 *     ws.send(msg);
 * }
 * 
 * function waitForSocketConnection(ws, callback) {
 *     setTimeout(
 *         function () {
 *             if (ws.readyState === 1) {
 *                 if(callback != null){
 *                     callback();
 *                 }
 *                 return;
 *             } else {
 *                 waitForSocketConnection(ws, callback);
 *             }
 *         }, 50);
 * }
 * 
 * var id = {{.Gadget.ID}};
 * var ready = false;
 * 
 * function sendCommand(id, info) {
 *     if (!ready) {
 *         showNotReady(id);
 *     }
 *     var cmd;
 *     if (document.getElementById(id).checked) {
 *         cmd = info.on[0];
 *     } else {
 *         cmd = info.off[0];
 *     }
 *     doSendComamnd(cmd);
 * }
 * 
 * waitForSocketConnection(ws, function() {
 *     ready = true;
 *     doSendComamnd("update");
 * });
 * 
 * ws.onmessage = function(message) {
 *     var msg = JSON.parse(message.data);
 *     if ((msg.type == "update" && msg.sender == "method runner") || msg.type == "method update") {
 *         showMethod(msg.method);
 *     } else if (msg.type == "update") {
 *         updateIO(msg);
 *     }
 * };
 *  */
{{end}}
