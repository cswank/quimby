'use strict';

angular.module('quimby.admin')
    .controller('UserCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', '$users', '$location', '$routeParams', function($scope, $rootScope, $gadgets, $auth, $users, $location, $routeParams) {
        $scope.editUser = {username: $routeParams.id};

        $rootScope.$watch('user', function(user) {
            if (user != {} && $scope.gadget != {} && $scope.editUser.username != "new-user") {
                $users.get($scope.editUser.username, function(data) {
                    console.log("got user", data);
                    $rootScope.links = [
                        {href:"#/admin", name:"admin"},
                        {href:"#/admin/users/" + $scope.editUser.username, name:data.name}
                    ];
                    $scope.editUser = data;
                });
            }
        });

        $scope.save = function() {
            
        };

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

