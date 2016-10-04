'use strict';

angular.module('quimby.filters', [])
    .filter('onoff', function() {
        return function(value) {
            var io = "";
            angular.forEach(value.io, function(value, key) {
                if (value == true) {
                    return io = "*";
                }
            });
            return value.value ? 'on' + io : 'off';
        };
    })
    .filter('onoffColor', function() {
        return function(input) {
            return input ? 'green' : 'red';
        };
    });
