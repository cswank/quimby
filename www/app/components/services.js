'use strict';

angular.module('quimby.services', [])
    .service('$gadgets', ['$http', function ($http) {
        var commands = {};
        var locations = {};
        this.getGadgets = function(name, callback) {
            $http.get("/api/gadgets/" +  name + "/values").success(function(data) {
                locations = data;
                callback(locations);
            });
            $http.get("/api/gadgets/" +  name + "/status").success(function(statuses) {
                angular.forEach(statuses, function(value, key) {
                    if (value.info.direction == "output") {
                        commands[key] = {on: value.info.on, off: value.info.off};
                    }
                });
            });
        }
        this.toggle =  function(location, name, callback) {
            var val = locations[location][name].value;
            var onoff = val ? "off":"on";
            var command = commands[location + " " + name][onoff];
            console.log(command);
            callback(command);
        }
        this.update = function(msg) {
            console.log(msg);
        }
    }])
    .factory('$sockets', ['$location', '$http', '$timeout', '$routeParams', function($location, $http, $timeout, $routeParams) {
        var ws;
        var outWs;
        var statusPromise;
        var host;
        var callback;
        
        function getWebsockets() {
            var prot = "wss";
            if ($location.protocol() == "http") {
                prot = "ws";
            }
            var url = prot + "://" + $location.host() + ":8111/api/gadgets/" + $routeParams.name + "/updates";
            ws = new WebSocket(url);
            console.log("got ws", ws);
            return ws;
        }

        function sendMessage(message) {
            message = JSON.parse(message);
            $timeout.cancel(statusPromise);
            var url = "/api/gadgets/" + host + "/commands";
            $http.post(url, message.message).success(function(data) {
                getStatus();
            });
        }
        
        function doConnect(errorCallback) {
            if(ws != undefined) {
                ws.close();
                ws = null;
            }
            ws = getWebsockets(host);
            ws.onopen = function() {
            };
            ws.onerror = function() {
            };
            ws.onmessage = function(message) {
                message = JSON.parse(message.data);
                console.log("got msg", message);
                callback(message);
            };
        }
        
        return {
            connect: function(cb) {
                callback = cb;
                doConnect();
            },
            send: function(command) {
                console.log("sending", command);
                ws.send(JSON.stringify({
                    type: "command",
                    body: command,
                }));
            },
            close: function() {
                if (ws != undefined) {
                    ws.close();
                }
            }
        }
    }]);
