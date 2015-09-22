'use strict';

angular.module('quimby.gadgets', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:name', {
            templateUrl: 'gadgets/gadgets.html',
            controller: 'GadgetsCtrl'
        });
    }])

.controller('GadgetsCtrl', ['$scope', '$gadgets', '$sockets', '$routeParams', function($scope, $gadgets, $sockets, $routeParams) {
    $scope.name = $routeParams.name;
    $gadgets.getGadgets($scope.name, function(locations) {
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
