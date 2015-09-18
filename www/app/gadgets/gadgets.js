'use strict';

angular.module('quimby.gadgets', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/gadgets', {
    templateUrl: 'gadgets/gadgets.html',
    controller: 'GadgetsCtrl'
  });
}])

.controller('GadgetsCtrl', ['$scope', function($scope) {
    $scope.gadgets = http.get("api/gadgets");
}]);
