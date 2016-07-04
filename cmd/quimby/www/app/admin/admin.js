'use strict';

angular.module('quimby.admin', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {

        $routeProvider.when('/admin/add-user', {
            templateUrl: 'admin/add-user.html',
            controller: 'AddUserCtrl'
        });
        
        $routeProvider.when('/admin', {
            templateUrl: 'admin/admin.html',
            controller: 'AdminListCtrl'
        });

        $routeProvider.when('/admin/new-gadget', {
            templateUrl: 'admin/gadget.html',
            controller: 'NewGadgetCtrl'
        });
        
        $routeProvider.when('/admin/:id', {
            templateUrl: 'admin/gadget.html',
            controller: 'AdminCtrl'
        });
        
        $routeProvider.when('/admin/users/:id', {
            templateUrl: 'admin/user.html',
            controller: 'UserCtrl'
        });

        
    }])

    .controller('NewGadgetCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', '$location', function($scope, $rootScope, $gadgets, $auth, $location) {
        $rootScope.links = [
            {href:"#/admin", name:"admin"}
        ];
        $scope.gadget = {
            name: "new gadget"
        };
        $scope.save = function() {
            $gadgets.save($scope.gadget, function(data) {
                $location.path("/admin");
            })
        }

        $scope.cancel = function() {
            $location.path("/admin");
        }
        
    }])

    .controller('AdminListCtrl', ['$scope', '$rootScope', '$gadgets', '$users', '$auth', function($scope, $rootScope, $gadgets, $users, $auth) {
        $rootScope.links = [
            {href:"#/admin", name:"admin"}
        ];
        $rootScope.$watch('user', function(user) {
            if (user != {}) {
                $gadgets.getGadgets(function(data) {
                    $scope.gadgets = data;
                });
                $users.getAll(function(data) {
                    $scope.users = data;
                });
            } else {
                $scope.gadgets = {};
            }
        });
    }]);
