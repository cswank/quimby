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

        this.save = function(user, callback) {
            var url = "/api/users";
            $http.post(url, user).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't save user");
            });
        }

        this.updatePermission = function(user, callback) {
            var url = "/api/users/" + user.username + "/permission";
            $http.post(url, user).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't save user");
            });
        }

        this.updatePassword = function(user, callback) {
            var url = "/api/users/" + user.username + "/password";
            $http.post(url, user).success(function(data) {
                callback(data);
            }).error(function() {
                console.log("didn't save user");
            });
        }
    }])

    .service('$gadgets', ['$http', function ($http) {
        var commands = {};
        var locations = {};
        var targets = {};
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
            targets = {};
            method = {};
            $http.get("/api/gadgets/" +  id + "/status").success(function(statuses) {
                var directions = {};
                angular.forEach(statuses, function(value, key) {
                    var v;
                    if (locations[value.location] == undefined) {
                        v = {};
                    } else {
                        v = locations[value.location];
                    }
                    v[value.name] = value.value;
                    locations[value.location] = v;

                    if (value.info.direction == "output") {
                        var t;
                        if (targets[value.location] == undefined) {
                            t = {};
                        } else {
                            t = targets[value.location];
                        }
                        t[value.name] = value.target_value;
                        targets[value.location] = t;
                    }
                    
                    if (value.info.direction != undefined) {
                        directions[key] = value.info.direction;
                    }
                    if (key == "method runner") {
                        method = value.method;
                    } else if (value.info.direction == "output") {
                        commands[key] = {on: value.info.on, off: value.info.off};
                    }
                });
                callback(locations, directions, targets, method);
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
