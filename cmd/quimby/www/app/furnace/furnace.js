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
            $scope.furnace = {};
            angular.copy(locations.home.furnace, $scope.furnace);
        });

        $scope.change = function(item) {
            var cmd;
            if (item == "heat") {
                if ($scope.locations.home.furnace.io.heat) {
                    cmd = "turn off furnace";
                } else {
                    cmd = "heat home to 75 F";
                }
            } else if (item == "cool") {
                if ($scope.locations.home.furnace.io.cool) {
                    cmd = "turn off furnace";
                } else {
                    cmd = "cool home to 70 F";
                }
            } else if (item == "fan") {
                if ($scope.locations.home.furnace.io.fan) {
                    //cmd = "turn off furnace";
                } else {
                    
                }
            }
            if (cmd) {
                $sockets.send(cmd);
            }
        };
        
        $sockets.connect(function(msg) {
            if (msg.type == "update") {
                $scope.$apply(function() {
                    $scope.locations[msg.location][msg.name] = msg.value;
                    if (msg.location == "home" && msg.name == "furnace") {
                        angular.copy(msg.value, $scope.furnace);
                    }
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
