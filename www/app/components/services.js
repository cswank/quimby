'use strict';

angular.module('quimby.services', [])
    .service('$gadgets', ['$http', function ($http) {
        var commands = {};
        var locations = {};
        this.getGadgets = function(callback) {
            $http.get("/api/gadgets").success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get gadgets");
            });
        }
        this.getDevices = function(name, callback) {
            $http.get("/api/gadgets/" +  name + "/values").success(function(data) {
                locations = data;
                callback(locations);
            }).error(function() {
              
            });
            $http.get("/api/gadgets/" +  name + "/status").success(function(statuses) {
                angular.forEach(statuses, function(value, key) {
                    if (value.info.direction == "output") {
                        commands[key] = {on: value.info.on, off: value.info.off};
                    }
                });
            });
        }
        this.send =  function(location, name, callback) {
            var val = locations[location][name].value;
            var onoff = val ? "off":"on";
            var command = commands[location + " " + name][onoff];
            callback(command);
        }
        this.update = function(msg) {
            
        }
    }])
    .factory('$sockets', ['$location', '$http', '$timeout', '$routeParams', function($location, $http, $timeout, $routeParams) {
        var ws;
        var outWs;
        var statusPromise;
        var host;
        
        function getWebsocket() {
            var prot = "wss";
            if ($location.protocol() == "http") {
                prot = "ws";
            }
            var url = prot + "://" + $location.host() + ":" + $location.port() + "/api/gadgets/" + $routeParams.name + "/websocket";
            ws = new WebSocket(url);
            return ws;
        }

        function doConnect(callback) {
            if(ws != undefined) {
                ws.close();
                ws = null;
            }
            ws = getWebsocket(host);
            ws.onopen = function() {
            };
            ws.onerror = function() {
            };
            ws.onmessage = function(message) {
                message = JSON.parse(message.data);
                callback(message);
            };
        }
        
        return {
            connect: function(cb) {
                doConnect(cb);
            },
            send: function(command) {
                var msg = JSON.stringify({
                    sender: "quimby",
                    type: "command",
                    body: command,
                });
                ws.send(msg);
            },
            close: function() {
                if (ws != undefined) {
                    ws.close();
                }
            }
        }
    }]);
