'use strict';

angular.module('quimby.admin')
    .controller('AddUserCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', '$users', '$location', '$routeParams', function($scope, $rootScope, $gadgets, $auth, $users, $location, $routeParams) {
        
        $scope.add = function() {
            
        }
            
        $scope.cancel = function() {
            $location.path("/admin");
        }
        
    }]);

