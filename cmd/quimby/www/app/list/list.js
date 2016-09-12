'use strict';

angular.module('quimby.list', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets', {
            templateUrl: 'list/list.html',
            controller: 'ListCtrl'
        });
    }])

.controller('ListCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', function($scope, $rootScope, $gadgets, $auth) {
    $rootScope.links = [];
    $rootScope.$watch('user', function(user) {
        if (user != {}) {
            $gadgets.getGadgets(function(data) {
                $scope.gadgets = _.filter(data, function(d) { return !d.disabled });
            });
        } else {
            $scope.gadgets = {};
        }
    });
}]);
