'use strict';

angular.module('quimby.gadget', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadget/:id', {
            templateUrl: '/gadget/gadget.html',
            controller: 'GadgetCtrl'
        });
    }])

    .controller('GadgetCtrl', ['$scope', '$gadgets', '$sockets', '$routeParams', '$mdSidenav', function($scope, $gadgets, $sockets, $routeParams, $mdSidenav, $rootScope) {
        
        
        $scope.method = {};
        $scope.id = $routeParams.id;
        $scope.decimals = 1;

        $gadgets.getGadget($scope.id, function(data) {
            $scope.gadget = data;
            $rootScope.links = [
                {href:"#/gadget" + id, name:data.name},
            ]
        });
        
        $gadgets.getDevices($scope.id, function(locations, directions, method) {
            $scope.directions = directions;
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
