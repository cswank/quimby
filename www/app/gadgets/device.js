'use strict';

angular.module('quimby.directives')
    .directive("device", ['$gadgets', '$sockets', function($gadgets, $sockets) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadgets/device.html?t=" + new Date().getTime(),
            scope: {
                name: '=',
                value: '=',
                location: '='
            },
            controller: function($scope) {
                $scope.toggle = function() {
                    $gadgets.send($scope.location, $scope.name, $sockets.send);
                }
            }
        }
    }]);
