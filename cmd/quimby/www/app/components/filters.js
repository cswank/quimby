'use strict';

angular.module('quimby.filters', [])
    .filter('onoff', function() {
        return function(input) {
            return input ? 'on' : 'off';
        };
    })
    .filter('onoffColor', function() {
        return function(input) {
            return input ? 'green' : 'red';
        };
    });
