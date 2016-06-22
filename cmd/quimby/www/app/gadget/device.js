'use strict';

angular.module('quimby.services')
    .directive("device", ['$gadgets', '$sockets', '$mdDialog', '$routeParams', '$location', function($gadgets, $sockets, $mdDialog, $routeParams, $location) {
        return {
            restrict: "E",
            replace: true,
            transclude: true,
            templateUrl: "/gadget/device.html?t=" + new Date().getTime(),
            scope: {
                name: '=',
                value: '=',
                location: '=',
                direction: '=',
                decimals: '='
            },
            controller: function($scope) {
                $scope.toggle = function() {
                    $gadgets.send($scope.location, $scope.name, $sockets.send);
                }
                $scope.showHistory = function(location, device) {
                    var url = "/gadgets/" + $routeParams.id + "/history/";
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
