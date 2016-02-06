'use strict';

angular.module('quimby.services')
    .directive("method", ["$sockets", "$methods", "$mdSidenav", "$localStorage", function($sockets, $methods, $mdSidenav, $localStorage) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadgets/method.html?t=" + new Date().getTime(),
            scope: {
                uuid: '=',
                method: '=',
                socket: '='
            },
            controller: function($scope) {
                $scope.$storage = $localStorage.$default({methods:[]});
                $scope.confirm = function(step) {
                    var msg = {
                        type: 'method update',
                        sender: 'client',
                        body:step
                    };
                    $sockets.update(msg);
                };
                $scope.checkUserPrompt = function(i) {
                    var step = $scope.method.steps[i];
                    return step != undefined && step.indexOf("wait for user") == 0 && i == $scope.method.step;
                };

                $scope.toggle = function() {
                    $mdSidenav('new-method').toggle();
                }

                $scope.close = function () {
                    $mdSidenav('new-method').close();
                };

                
                $scope.runStoredMethod = function () {
                    $mdSidenav('new-method').close();
                    $methods.runMethod($scope.uuid, $scope.$storage.methods[$scope.selected]);
                    $scope.selected = -1;
                }

                $scope.deleteStoredMethod = function () {
                    $mdSidenav('new-method').close();
                    $scope.$storage.methods.splice($scope.selected, 1);
                    $scope.selected = -1;
                }

                $scope.run = function () {
                    $mdSidenav('new-method').close();
                    var method = $scope.newMethod.split("\n")
                    $methods.runMethod($scope.uuid, method);
                    $scope.$storage.methods.push(method);
                };
            }
        }
    }]);

angular.module('quimby.services')
    .service('$methods', ['$http', function ($http) {
        this.runMethod = function(id, method) {
            $http.post(
                "/api/gadgets/" + id + "/method",
                {method: method}
            )
        }
    }]);

angular.module('quimby.filters')
    .filter('countdown', [function() {
        return function(input) {
            var s = input % 60;
            s = (s < 10) ? '0' + s : s;
            var m = Math.floor(input / 60);
            m = (m < 10) ? '0' + m : m;
            var h = Math.floor(input / 3600);
            return h + ':' + m + ':' + s;
        }
    }]);
