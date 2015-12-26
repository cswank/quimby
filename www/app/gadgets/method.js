'use strict';

angular.module('quimby.services')
    .directive("method", ["$sockets", function($sockets) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadgets/method.html?t=" + new Date().getTime(),
            scope: {
                method: '=',
                socket: '='
            },
            controller: function($scope) {

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
            }
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
