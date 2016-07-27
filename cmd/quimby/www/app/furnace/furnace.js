'use strict';

angular.module('quimby.furnace', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadget/furnace/:id', {
            templateUrl: '/furnace/furnace.html',
            controller: 'FurnaceCtrl'
        });
    }])

    .controller('FurnaceCtrl', ['$scope', '$rootScope', '$gadgets', '$sockets', '$routeParams', function($scope, $rootScope, $gadgets, $sockets, $routeParams) {
        
        
        $scope.method = {};
        $scope.id = $routeParams.id;
        $scope.decimals = 1;

        $gadgets.getGadget($scope.id, function(data) {
            $scope.gadget = data;
            $rootScope.links = [
                {href:"#/gadget" + $scope.id, name:data.name},
            ]
        });
        
        $gadgets.getDevices($scope.id, function(locations, directions, method) {
            $scope.locations = locations;
            $scope.method = method;
        });

        $scope.change = function(item) {
            var cmd;
            if (item == "heat") {
                if ($scope.locations.home.furnace.io.heat) {
                    cmd = "heat home to 75 F";
                } else {
                    cmd = "turn off furnace";
                }
            } else if (item == "cool") {
                if ($scope.locations.home.furnace.io.cool) {
                    cmd = "cool home to 70 F";
                } else {
                    cmd = "turn off furnace";
                }
            } else if (item == "fan") {
                if ($scope.locations.home.furnace.io.fan) {
                    //cmd = "turn off furnace";
                } else {
                    
                }
            }
            console.log("cmd", cmd);
            if (cmd) {
                $sockets.send(cmd);
            }
        };
        
        $sockets.connect(function(msg) {
            console.log("update", msg);
            if (msg.type == "update") {
                $scope.$apply(function() {
                    $scope.locations[msg.location][msg.name] = msg.value;
                })
            } else if (msg.type == "method update") {
                $scope.$apply(function() {
                    $scope.method = msg.method;
                })
            }
        });
        
        $scope.$on('$locationChangeStart', function( event ) {
            $sockets.close();
        });
    }]);
