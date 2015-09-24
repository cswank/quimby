'use strict';

angular.module('quimby.gadgets', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets', {
            templateUrl: 'gadgets/gadgets.html',
            controller: 'GadgetsCtrl'
        });
    }])

.controller('GadgetsCtrl', ['$scope', '$gadgets', '$auth', function($scope, $gadgets, $auth) {
    $auth.getUser(function(user) {
        $gadgets.getGadgets(function(data) {
            $scope.gadgets = data;
        });
    });
}]);
