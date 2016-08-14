'use strict';

angular.module('quimby.services', [])
    .service('$users', ['$http', function ($http) {
        this.getAll = function(callback) {
            $http.get("/api/users").success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get users");
            });
        }
        
        this.get = function(username, callback) {
            $http.get("/api/users/" + username).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get users");
            });
        }

        this.delete = function(username, callback) {
            $http.delete("/api/users/" + username).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get users");
            });
        }
    }])

    .service('$gadgets', ['$http', function ($http) {
        var commands = {};
        var locations = {};
        var method = {};
        this.getGadgets = function(callback) {
            $http.get("/api/gadgets").success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get gadgets");
            });
        }
        this.getGadget = function(id, callback) {
            $http.get("/api/gadgets/" + id).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get gadget");
            });
        }
        this.getDevice = function(id, callback) {
            $http.get("/api/gadgets/" + id).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get gadget");
            });
        }
        this.save = function(gadget, callback) {
            $http.post("/api/gadgets", gadget).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get gadget");
            });
        }
        this.delete = function(id, callback) {
            $http.delete("/api/gadgets/" + id).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't get gadget");
            });
        }
        this.update = function(gadget, callback) {
            $http.post("/api/gadgets/" + gadget.id, gadget).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't save gadget");
            });
        }
        this.getStatus = function(id, callback) {
            $http.get("/api/gadgets/" +  id + "/status").success(function(statuses) {
                callback(statuses);
            });
        }
        this.getDevices = function(id, callback) {
            commands = {};
            locations = {};
            method = {};
            $http.get("/api/gadgets/" +  id + "/values").success(function(data) {
                locations = data;
                $http.get("/api/gadgets/" +  id + "/status").success(function(statuses) {
                    var directions = {};
                    angular.forEach(statuses, function(value, key) {
                        if (value.info.direction != undefined) {
                            directions[key] = value.info.direction;
                        }
                        if (key == "method runner") {
                            method = value.method;
                        } else if (value.info.direction == "output") {
                            commands[key] = {on: value.info.on, off: value.info.off};
                        }
                    });
                    callback(locations, directions, method);
                });
            }).error(function() {
              
            });
            
        }
        
        this.send =  function(location, name, callback) {
            var val = locations[location][name].value;
            var onoff = val ? "off":"on";
            var command = commands[location + " " + name][onoff][0];
            callback(command);
        }

        this.sendWithArgs =  function(location, name, args, callback) {
            var val = locations[location][name].value;
            var onoff = val ? "off":"on";
            var command = commands[location + " " + name][onoff][0];
            command += " " + args;
            callback(command);
        }

        this.getCommands =  function(location, name, callback) {
            var val = locations[location][name].value;
            var onoff = val ? "off":"on";
            callback(commands[location + " " + name][onoff]);
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
            var url = prot + "://" + $location.host() + ":" + $location.port() + "/api/gadgets/" + $routeParams.id + "/websocket";
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
            update: function(msg) {
                ws.send(JSON.stringify(msg));
            },
            close: function() {
                if (ws != undefined) {
                    ws.close();
                }
            }
        }
    }]);
