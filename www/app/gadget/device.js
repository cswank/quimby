'use strict';

angular.module('quimby.services')
    .service('$stats', ['$http', function ($http) {
        this.getStats = function(id, location, name, callback) {
            var url = "/api/gadgets/" + id + "/locations/" + location + "/devices/" + name + "/datapoints";
            $http.get(url).success(function(data) {
                callback(data);
            })
        }
    }])
    .directive("device", ['$gadgets', '$sockets', '$mdDialog', '$routeParams', '$stats', '$location', function($gadgets, $sockets, $mdDialog, $routeParams, $stats, $location) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadget/device.html?t=" + new Date().getTime(),
            scope: {
                name: '=',
                value: '=',
                location: '=',
                direction: '='
            },
            controller: function($scope) {
                $scope.toggle = function() {
                    $gadgets.send($scope.location, $scope.name, $sockets.send);
                }
                $scope.showGraph = function(location, device) {
                    var url = "/graphs/" + $routeParams.id;
                    console.log(url);
                    $location.path(url).search({location: location, device: device});
                }
            }
        }
    }])
    .directive('ngRightClick', [function($parse) {
        return function(scope, element, attrs) {
            var fn = $parse(attrs.ngRightClick);
            element.bind('contextmenu', function(event) {
                scope.$apply(function() {
                    event.preventDefault();
                    fn(scope, {$event:event});
                });
            });
        }
    }]);
