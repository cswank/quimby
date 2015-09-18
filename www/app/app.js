'use strict';

// Declare app level module which depends on views, and components
angular.module('quimby', [
  'ngRoute',
  'quimby.gadgets'
]).
config(['$routeProvider', function($routeProvider) {
  $routeProvider.otherwise({redirectTo: '/gadgets'});
}]);
