'use strict';

function DeleteGadgetController($scope, $mdDialog, name) {
    $scope.name = name;
    $scope.cancel = function() {
        $mdDialog.cancel();
    };
    $scope.del = function() {
        $mdDialog.hide(true);
    };
}


angular.module('quimby.admin')
    .controller('GadgetAdminCtrl', ['$scope', '$rootScope', '$gadgets', '$auth', '$routeParams', '$location', '$mdDialog', function($scope, $rootScope, $gadgets, $auth, $routeParams, $location, $mdDialog) {
        var id = $routeParams.id;
        $scope.gadget = {};
        
        $rootScope.links = [];
        
        $rootScope.$watch('user', function(user) {
            if (user != {} && $scope.gadget != {}) {
                $gadgets.getGadget(id, function(data) {
                    $rootScope.links = [
                        {href:"#/admin", name:"admin"},
                        {href:"#/admin/" + id, name:data.name}
                    ];
                    $scope.gadget = data;
                });
            }
        });

        $scope.interfaces = [{name:"default"}, {name:"furnace"}];

        $scope.save = function() {
            $gadgets.update($scope.gadget, function(data) {
                $location.path("/admin");
            });
        }
            
        $scope.cancel = function() {
            $location.path("/admin");
        }

        $scope.delete = function($mdOpenMenu, ev) {
            $mdDialog.show({
                controller: DeleteGadgetController,
                templateUrl: 'admin/confirm.html?t=' + new Date().getTime(),
                locals: {
                    name: $scope.gadget.name
                },
                targetEvent: ev
            }).then(function(result) {
                if (result == true) {
                    $gadgets.delete(id, function() {
                        $location.path("/admin");  
                    });
                }
            });
        }
    }]);

