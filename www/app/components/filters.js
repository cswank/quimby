'use strict';

angular.module('quimby.filters', [])
    .filter('onoff', function() {
        return function(input) {
            return input ? 'on' : 'off';
        };
    });
