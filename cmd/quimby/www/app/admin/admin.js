'use strict';

angular.module('quimby.admin', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
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

    .controller('AdminListCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', function($scope, $rootScope, $gadgets, $auth) {
        $rootScope.links = [
            {href:"#/admin", name:"admin"}
        ];
        $rootScope.$watch('user', function(user) {
            if (user != {}) {
                $gadgets.getGadgets(function(data) {
                    $scope.gadgets = data;
                });
            } else {
                $scope.gadgets = {};
            }
        });
    }]);
