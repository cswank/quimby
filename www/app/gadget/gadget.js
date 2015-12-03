'use strict';

angular.module('quimby.gadget', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:id', {
            templateUrl: 'gadget/gadget.html',
            controller: 'GadgetCtrl'
        });
    }])

    .controller('GadgetCtrl', ['$scope', '$gadgets', '$sockets', '$routeParams', function($scope, $gadgets, $sockets, $routeParams) {
        $scope.id = $routeParams.id;

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
            }
        });
        
        $scope.$on('$locationChangeStart', function( event ) {
            $sockets.close();
        });

    }]);
