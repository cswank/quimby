'use strict';

angular.module('quimby.services')
    .directive("device", ['$gadgets', '$sockets', '$mdDialog', '$routeParams', '$location', function($gadgets, $sockets, $mdDialog, $routeParams, $location) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadgets/device.html?t=" + new Date().getTime(),
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
