'use strict';

// Declare app level module which depends on views, and components
angular.module('quimby', [
    'ngRoute',
    'ngMaterial',
    'ngAnimate',
    'ngStorage',
    'quimby.gadgets',
    'quimby.gadget',    
    'quimby.directives',
    'quimby.services',
    'quimby.filters'
]).
    config(['$routeProvider', function($routeProvider) {
        $routeProvider.otherwise({redirectTo: '/gadgets'});
    }]);
