'use strict';

angular.module('quimby.directives')
    .directive("gadget", ['$gadgets', '$sockets', function($gadgets, $sockets) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/components/gadget/gadget.html?t=" + new Date().getTime(),
            controller: function($scope) {
                $scope.toggle = function(location, name) {
                    $gadgets.toggle(location, name, $sockets.send);
                }
                $sockets.connect($gadgets.update);
            }
        }
    }]);
        
