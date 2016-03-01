'use strict';

// Declare app level module which depends on views, and components
angular.module('quimby', [
    'ngRoute',
    'ngMaterial',
    'ngAnimate',
    "angularMoment",
    'ngStorage',
    'nvd3ChartDirectives',
    'quimby.gadget',
    'quimby.list',
    'quimby.history',
    'quimby.directives',
    'quimby.services',
    'quimby.filters'
]).
    config(['$routeProvider', function($routeProvider) {
        $routeProvider.otherwise({redirectTo: '/gadgets'});
    }]);
