'use strict';

angular.module('quimby.admin')
    .controller('UserCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', '$location', function($scope, $rootScope, $gadgets, $auth, $location) {
        var username = $routeParams.id;

        $rootScope.$watch('user', function(user) {
            if (user != {} && $scope.gadget != {}) {
                $users.getUser(username, function(data) {
                    $rootScope.links = [
                        {href:"#/admin", name:"admin"},
                        {href:"#/admin/users" + username, name:data.name}
                    ];
                    $scope.user = data;
                });
            }
        });

        $scope.delete = function() {
            $mdDialog.show({
                controller: DeleteGadgetController,
                templateUrl: 'admin/confirm.html?t=' + new Date().getTime(),
                locals: {
                    name: $scope.user.username
                },
                targetEvent: ev
            }).then(function(result) {
                if (result == true) {
                    $users.delete(username, function() {
                        $location.path("/admin");  
                    });
                }
            });
        }

        $scope.cancel = function() {
            $location.path("/admin");
        }
        
    }]);

