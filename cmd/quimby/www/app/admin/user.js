'use strict';

function UpdatePasswordController($scope, $mdDialog, username) {
    $scope.password = {first:"", second: ""};
    $scope.username = username;
    $scope.cancel = function() {
        $mdDialog.cancel();
    };
    $scope.save = function() {
        if ($scope.password.first != $scope.password.second) {
            $scope.password.first = "";
            $scope.password.second = "";
            $scope.passwordError = "passwords don't match";
        } else {
            $mdDialog.hide($scope.password.first);
        }
    };
}


angular.module('quimby.admin')
    .controller('UserCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', '$users', '$location', '$routeParams', '$mdDialog', function($scope, $rootScope, $gadgets, $auth, $users, $location, $routeParams, $mdDialog) {
        $scope.editUser = {username: $routeParams.id};
        $scope.password = {};
        $scope.isNew = false;
        if ($scope.editUser.username == "new-user") {
            $scope.isNew = true;
            $scope.editUser.permission = "read";
        }

        $rootScope.$watch('user', function(user) {
            if (user != {} && $scope.gadget != {} && $scope.editUser.username != "new-user") {
                $users.get($scope.editUser.username, function(data) {
                    $rootScope.links = [
                        {href:"#/admin", name:"admin"},
                        {href:"#/admin/users/" + $scope.editUser.username, name:data.name}
                    ];
                    $scope.editUser = data;
                });
            }
        });

        $scope.done = function() {
            $location.path("/admin");
        };

        $scope.save = function() {
            if ($scope.isNew) {
                if ($scope.password.first != $scope.password.second) {
                    $scope.password.first = "";
                    $scope.password.second = "";
                    $scope.passwordError = "passwords don't match";
                    return;
                } else {
                    $scope.editUser.password = $scope.password.first;
                }
                
                $users.save($scope.editUser, function(data) {
                    $scope.qrData = data;
                });
            } else {
                $users.updatePermission($scope.editUser, function() {
                    $location.path("/admin");
                });
            }
        };

        $scope.updatePassword = function(ev) {
            $mdDialog.show({
                controller: UpdatePasswordController,
                templateUrl: 'admin/password.html?t=' + new Date().getTime(),
                locals: {username: $scope.editUser.username},
                targetEvent: ev
            }).then(function(pw) {
                if (pw != "") {
                    $scope.editUser.password = pw;
                    $users.updatePassword($scope.editUser, function() {
                        $location.path("/admin");
                    });
                }
            });
        }

        $scope.delete = function(ev) {
            $mdDialog.show({
                controller: DeleteGadgetController,
                templateUrl: 'admin/confirm.html?t=' + new Date().getTime(),
                locals: {
                    name: $scope.editUser.username
                },
                targetEvent: ev
            }).then(function(result) {
                if (result == true) {
                    $users.delete($scope.editUser.username, function() {
                        $location.path("/admin");
                    });
                }
            });
        }

        $scope.cancel = function() {
            $location.path("/admin");
        }
        
    }]);

