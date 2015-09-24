'use strict';

angular.module('quimby.gadget', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:name', {
            templateUrl: 'gadget/gadget.html',
            controller: 'GadgetCtrl'
        });
    }])

.controller('GadgetCtrl', ['$scope', '$gadgets', '$sockets', '$routeParams', function($scope, $gadgets, $sockets, $routeParams) {
    $scope.name = $routeParams.name;
    $gadgets.getDevices($scope.name, function(locations) {
        $scope.locations = locations;
    });
    
    $sockets.connect(function(msg) {
        if (msg.type == "update") {
            $scope.$apply(function() {
                $scope.locations[msg.location][msg.name] = msg.value;
            })
        }
    });
}]);
