'use strict';

// Declare app level module which depends on views, and components
angular.module('quimby', [
    'ngRoute',
    'ngMaterial',
    'ngAnimate',
    'quimby.gadgets',
    'quimby.directives',
    'quimby.services'
]).
    config(['$routeProvider', function($routeProvider) {
        $routeProvider.otherwise({redirectTo: '/gadgets'});
    }]);
