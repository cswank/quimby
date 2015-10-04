'use strict';

angular.module('quimby.directives')
    .directive("device", ['$gadgets', '$sockets', function($gadgets, $sockets) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadget/device.html?t=" + new Date().getTime(),
            scope: {
                name: '=',
                value: '=',
                location: '=',
                direction: '='
            },
            controller: function($scope) {
                $scope.toggle = function() {
                    $gadgets.send($scope.location, $scope.name, $sockets.send);
                }
            }
        }
    }]);
