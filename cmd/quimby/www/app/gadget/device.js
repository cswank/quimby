'use strict';

function ArgsController($scope, $mdDialog, command) {

    $scope.args = "";
    $scope.command = command;

    $scope.cancel = function() {
        $mdDialog.cancel();
    };
    $scope.send = function() {
        var cmd = $scope.command + " " + $scope.args;
        $mdDialog.hide(cmd);
    };
}

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
                $scope.args = function(ev) {
                    $gadgets.getCommand($scope.location, $scope.name, function(cmd) {
                        $mdDialog.show({
                            controller: ArgsController,
                            templateUrl: '/gadget/args.html?t=' + new Date().getTime(),
                            targetEvent: ev,
                            locals: {
                                command: cmd
                            },
                        }).then(function(args) {
                            if (args) {
                                $gadgets.sendWithArgs($scope.location, $scope.name, args, $sockets.send);
                            }
                        });
                    })
                }
                
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
