'use strict';

angular.module('quimby.admin')
    .controller('UserCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', '$users', '$location', '$routeParams', function($scope, $rootScope, $gadgets, $auth, $users, $location, $routeParams) {
        var username = $routeParams.id;

        console.log("user", username);
        $rootScope.$watch('user', function(user) {
            if (user != {} && $scope.gadget != {}) {
                $users.get(username, function(data) {
                    console.log("got user", data);
                    $rootScope.links = [
                        {href:"#/admin", name:"admin"},
                        {href:"#/admin/users" + username, name:data.name}
                    ];
                    $scope.editUser = data;
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

