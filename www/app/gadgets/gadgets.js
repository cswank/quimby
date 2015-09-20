'use strict';

angular.module('quimby.gadgets', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:name', {
            templateUrl: 'gadgets/gadgets.html',
            controller: 'GadgetsCtrl'
        });
    }])

.controller('GadgetsCtrl', ['$scope', '$gadgets', '$routeParams', function($scope, $gadgets, $routeParams) {
    $scope.name = $routeParams.name;
    $gadgets.getGadgets($scope.name, function(locations) {
        $scope.locations = locations;
    });
}]);
