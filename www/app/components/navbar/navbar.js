'use strict';

function LoginController($scope, $mdDialog, message) {
    $scope.message = message;
    $scope.user = {
        'username': '',
        'password': ''
    };
    $scope.cancel = function() {
        $mdDialog.cancel();
    };
    $scope.login = function() {
        $mdDialog.hide($scope.user);
    };
}

function LogoutController($scope, $mdDialog) {
    $scope.cancel = function() {
        $mdDialog.cancel();
    };
    $scope.logout = function() {
        $mdDialog.hide(true);
    };
}

angular.module('quimby.directives')
    .directive("navbar", ['$location', '$auth', '$mdDialog', '$mdSidenav', function($location, $auth, $mdDialog, $mdSidenav) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/components/navbar/navbar.html?t=" + new Date().getTime(),
            controller: function($scope, $rootScope, $timeout) {
                $scope.loggedIn = false;
                $scope.showLogout = function(ev) {
                    $mdDialog.show({
                        controller: LogoutController,
                        templateUrl: '/components/navbar/logout.html?t=' + new Date().getTime(),
                        targetEvent: ev,
                    }).then(function(result) {
                        if (result) {
                            $auth.logout(function() {
                                $mdSidenav('left').toggle();
                                $scope.user = {};
                                $scope.loggedIn = false;
                                $scope.requests = [];
                            });
                        }
                    });
                }
                $scope.showLogin = function(ev) {
                    $mdDialog.show({
                        controller: LoginController,
                        templateUrl: 'components/navbar/login.html?t=' + new Date().getTime(),
                        locals: {
                            message: $scope.message
                        },
                        targetEvent: ev
                    }).then(function(result) {
                        if (result == "forgot") {
                            $location.path("/reset-password");
                        } else if (result) {
                            $auth.login(result.username, result.password, function(user) {
                                $scope.message = "";
                                $scope.user = user;
                                $scope.loggedIn = true;
                                $location.path("/gadgets");
                            }, function(){
                                $scope.message = "the username or password is not correct, please try again";
                                $scope.showLogin();
                            });
                        }
                    });
                };
                
                $auth.getUser(function(user) {
                    if (user) {
                        $scope.user = user;
                        $scope.loggedIn = true;
                    } else {
                        $scope.user = {};
                    }
                });
            }
        }
    }])
