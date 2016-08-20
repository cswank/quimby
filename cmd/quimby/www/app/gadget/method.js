'use strict';

angular.module('quimby.services')
    .directive("method", ["$sockets", "$methods", "$brew", "$mdSidenav", "$localStorage", function($sockets, $methods, $brew, $mdSidenav, $localStorage) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadget/method.html?t=" + new Date().getTime(),
            scope: {
                uuid: '=',
                method: '=',
                socket: '='
            },
            controller: function($scope) {
                $scope.$storage = $localStorage.$default({methods:{}});
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
                    $scope.newMethod = {
                        title: $scope.selected,
                        steps: $scope.$storage.methods[$scope.uuid][$scope.selected].join("\n")
                    };
                    $scope.selected = "";
                }

                $scope.deleteStoredMethod = function () {
                    $mdSidenav('new-method').close();
                    delete $scope.$storage.methods[$scope.uuid][$scope.selected];
                    $scope.selected = "";
                }

                $scope.fetchBrewMethod = function() {
                    $brew.fetchMethod($scope.brew, function(data) {
                        $scope.newMethod = {
                            title: $scope.brew.name,
                            steps: data.join("\n")
                        };
                    })
                }

                $scope.run = function () {
                    $mdSidenav('new-method').close();
                    var steps = $scope.newMethod.steps.split("\n");
                    $methods.runMethod($scope.uuid, steps);
                    if ($scope.$storage.methods[$scope.uuid] == undefined) {
                        $scope.$storage.methods[$scope.uuid] = {};
                    }
                    if ($scope.newMethod.title != undefined) {
                        $scope.$storage.methods[$scope.uuid][$scope.newMethod.title] = steps;
                    }
                    $scope.newMethod = {};
                };
            }
        }
    }]);

angular.module('quimby.services')
    .service('$brew', ['$http', function ($http) {
        this.fetchMethod = function(brew, callback) {
            $http.get(
                "/api/beer/" + brew.name,
                {grain_temperature: brew.temperature}
            ).success(function(data) {
                callback(data);
            })
        }
    }])
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
