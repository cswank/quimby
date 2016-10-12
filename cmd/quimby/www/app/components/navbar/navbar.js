'use strict';

function LoginController($scope, $mdDialog, message) {
    $scope.message = message;
    $scope.user = {
        'username': '',
        'password': '',
        'tfa':''
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
            controller: function($scope, $rootScope, $timeout, $mdDialog) {
                var originatorEv;
                $scope.loggedIn = false;
                $rootScope.user = {};
                $scope.openMenu = function($mdOpenMenu, ev) {
                    originatorEv = ev;
                    $mdOpenMenu(ev);
                };
                $scope.showLogout = function(ev) {
                    $mdDialog.show({
                        controller: LogoutController,
                        templateUrl: '/components/navbar/logout.html?t=' + new Date().getTime(),
                        targetEvent: ev,
                    }).then(function(result) {
                        if (result) {
                            $auth.logout(function() {
                                $rootScope.user = {};
                                $scope.user = {};
                                $scope.loggedIn = false;
                                $location.path("/gadgets");
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
                            $auth.login(result.username, result.password, result.tfa, function(user) {
                                $scope.message = "";
                                $scope.user = user;
                                $scope.loggedIn = true;
                                $location.path("/gadgets");
                            }, function(data) {
                                console.log("not logged in", data);
                                $scope.message = "the username or password is not correct, please try again";
                                $scope.showLogin();
                            });
                        }
                    });
                };

                $scope.admin = function() {
                    $location.path('/admin');
                }
                
                $auth.getUser(function(user) {
                    if (user) {
                        $rootScope.user = user;
                        $scope.user = user;
                        $scope.loggedIn = true;
                    } else {
                        $scope.user = {};
                    }
                });
            }
        }
    }])
