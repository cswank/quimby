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

        $scope.doneSliding = function() {
            if ($scope.mode == "off") {
                return
            }
            $scope.change();
        }
        
        $gadgets.getStatus($scope.id, function(statuses) {
            var furnace = statuses["home furnace"];
            if (furnace.value.value) {
                if (furnace.value.command == "cool home") {
                    $scope.mode = "cool";
                } else if (furnace.value.command == "heat home") {
                    $scope.mode = "heat";
                }
            } else {
                $scope.mode = "off";
            }
            
            $scope.temperature = statuses["home temperature"].value.value;

            if (furnace.target_value != undefined) {
                $scope.target = statuses["home furnace"].target_value.value;
            } else {
                $scope.target = Math.floor($scope.temperature);
            }
        });

        $scope.change = function() {
            var cmd = "turn off furnace";
            if ($scope.mode == "heat") {
                cmd = "heat home to " + $scope.target + " F";
            } else if ($scope.mode == "cool") {
                cmd = "cool home to " + $scope.target + " F";
            }
            $sockets.send("turn off furnace");
            $sockets.send(cmd);
        };
        
        $sockets.connect(function(msg) {
            if (msg.type == "update") {
                $scope.$apply(function() {
                    if (msg.name == "temperature") {
                        $scope.temperature = msg.value.value;
                    }
                });
            }
        });
        
        $scope.$on('$locationChangeStart', function( event ) {
            $sockets.close();
        });
    }]);
