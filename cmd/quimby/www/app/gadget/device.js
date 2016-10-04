'use strict';

function ArgsController($scope, $mdDialog, commands) {

    $scope.args = "";
    $scope.commands = commands;

    if (commands.length == 1) {
        $scope.command = commands[0];
    }

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
                    $gadgets.getCommands($scope.location, $scope.name, function(cmds) {
                        $mdDialog.show({
                            controller: ArgsController,
                            templateUrl: '/gadget/args.html?t=' + new Date().getTime(),
                            targetEvent: ev,
                            locals: {
                                commands: cmds
                            },
                        }).then(function(cmd) {
                            if (cmd) {
                                $sockets.send(cmd);
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
