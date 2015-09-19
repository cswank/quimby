'use strict';

angular.module('quimby.gadgets', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:name', {
            templateUrl: 'gadgets/gadgets.html',
            controller: 'GadgetsCtrl'
        });
    }])

.controller('GadgetsCtrl', ['$scope', '$http', '$routeParams', function($scope, $http, $routeParams) {
    var commands = {};
    $scope.name = $routeParams.name;
    $http.get("/api/gadgets/" +  $scope.name + "/values").success(function(locations) {
        $scope.locations = locations;
    });

    $http.get("/api/gadgets/" +  $scope.name + "/status").success(function(statuses) {
        angular.forEach(statuses, function(value, key) {
            if (value.info.direction == "output") {
                commands[key] = {on: value.info.on, off: value.info.off};
            }
        });
    });

    $scope.toggle = function(location, name) {
        var val = $scope.locations[location][name].value;
        var key = location + " " + name;
        var onoff;
        if (val) {
            onoff = "off"
        } else {
            onoff = "on"
        }
        var command = commands[key][onoff];
        console.log(command);
    }
    
}]);
