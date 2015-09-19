'use strict';

angular.module('quimby.gadgets', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:name', {
            templateUrl: 'gadgets/gadgets.html',
            controller: 'GadgetsCtrl'
        });
    }])

.controller('GadgetsCtrl', ['$scope', '$http', '$routeParams', function($scope, $http, $routeParams) {
    $scope.name = $routeParams.name;
    $http.get("/api/gadgets/" +  $scope.name + "/values").success(function(locations) {
        $scope.locations = locations;
    });
}]);
