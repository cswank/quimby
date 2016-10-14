'use strict';

angular.module('quimby.filters', [])
    .filter('onoff', function() {
        return function(value, target) {
            var io = "";
            angular.forEach(value.io, function(value, key) {
                if (value == true) {
                    io = "*";
                }
            });
            
            var t = "";
            if (target != undefined) {
                t = " (" + target.value + ")";
            }
            return value.value ? 'on' + io + t : 'off';
        };
    })
    .filter('onoffColor', function() {
        return function(input) {
            return input ? 'green' : 'red';
        };
    })
    .filter('ioColor', function() {
        return function(value) {
            return value ? 'green' : 'black';
        };
    });
