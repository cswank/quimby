'use strict';

// Declare app level module which depends on views, and components
angular.module('quimby', [
    'ngRoute',
    'ngMaterial',
    'ngAnimate',
    "angularMoment",
    'ngStorage',
    'chart.js',
    'quimby.gadgets',
    'quimby.gadget',
    'quimby.graph',
    'quimby.directives',
    'quimby.services',
    'quimby.filters'
]).
    config(['$routeProvider', function($routeProvider) {
        $routeProvider.otherwise({redirectTo: '/gadgets'});
    }]);
