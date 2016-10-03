'use strict';

angular.module('quimby.gadget', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadget/default/:id', {
            templateUrl: '/gadget/gadget.html',
            controller: 'GadgetCtrl'
        });
    }])

    .controller('GadgetCtrl', ['$scope', '$rootScope', '$gadgets', '$sockets', '$routeParams', '$mdSidenav', function($scope, $rootScope, $gadgets, $sockets, $routeParams, $mdSidenav) {
        
        
        $scope.method = {};
        $scope.id = $routeParams.id;
        $scope.decimals = 1;

        $gadgets.getGadget($scope.id, function(data) {
            $scope.gadget = data;
            $rootScope.links = [
                {href:"#/gadget" + $scope.id, name:data.name},
            ]
        });
        
        $gadgets.getDevices($scope.id, function(locations, directions, targets, method) {
            $scope.directions = directions;
            $scope.targets = targets;
            $scope.locations = locations;
            $scope.method = method;
        });
        
        $sockets.connect(function(msg) {
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

        $scope.toggle = function() {
            $mdSidenav('left').toggle();
        }

        $scope.close = function () {
            $mdSidenav('left').close();
        };

    }]);
