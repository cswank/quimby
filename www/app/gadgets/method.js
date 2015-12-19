'use strict';

angular.module('quimby.services')
    .directive("method", [function() {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadgets/method.html?t=" + new Date().getTime(),
            scope: {
                method: '=',
            },
            controller: function($scope) {}
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
