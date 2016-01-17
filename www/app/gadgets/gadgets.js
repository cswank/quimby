'use strict';

angular.module('quimby.gadget', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:id', {
            templateUrl: '/gadgets/gadgets.html',
            controller: 'GadgetsCtrl'
        });
    }])

    .controller('GadgetsCtrl', ['$scope', '$gadgets', '$sockets', '$routeParams', '$mdSidenav', function($scope, $gadgets, $sockets, $routeParams, $mdSidenav) {
        $scope.method = {};
        $scope.id = $routeParams.id;
        $scope.decimals = 1;

        $gadgets.getGadget($scope.id, function(data) {
            $scope.gadget = data;
        });
        
        $gadgets.getDevices($scope.id, function(locations, directions) {
            $scope.directions = directions;
            $scope.locations = locations;
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
