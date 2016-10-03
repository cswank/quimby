'use strict';

angular.module('quimby.history', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:id/history', {
            templateUrl: 'history/history.html',
            controller: 'HistoryCtrl'
        });
    }])
    .controller('HistoryCtrl', ['$scope', '$routeParams', '$stats', '$gadgets', '$window', '$mdSidenav', function($scope, $routeParams, $stats, $gadgets, $window, $mdSidenav) {
        $scope.spans = {
            hour: 1,
            day: 24,
            week: 7 * 24,
            month: 7 * 24 * 30
        };

        $scope.summarize = 0;
        
        $scope.spanLabels = ["hour", "day", "week", "month"];
        
        $scope.selectedSpan = "hour";
        
        
        $scope.label = $routeParams.location + " " + $routeParams.device;
        $scope.id = $routeParams.id;

        $gadgets.getGadget($scope.id, function(data) {
            $scope.gadget = data;
        });

        function getEnd() {
            return encodeURIComponent(moment().utc().format());
        }

        function getStart() {
            return encodeURIComponent(moment().utc().subtract($scope.spans[$scope.selectedSpan], "hours").format());
        }

        $scope.data = [];
        
        $scope.getDate = function(){
            return function(d){
                return d3.time.format('%x %X')(new Date(d));  //uncomment for date format
            }
        };

        $scope.isSelected = function(key, x) {
            return _.findIndex($scope.data, function(item) {
                return item.key == key;
            }) > -1;
        };

        $scope.getHeight = function() {
            return 2 * $window.innerHeight / 3;
        }

        $scope.getSpanStyle = function(name) {
            if (name == $scope.selectedSpan) {
                return {color: '#3F51B5'};
            }
            return {};
        };
                
        
        $scope.getData = function(key) {
            $stats.getStats($scope.id, key, $scope.spans[$scope.selectedSpan], $scope.summarize, function(data) {
                var i = _.findIndex($scope.data, function(item) {
                    return item.key == key;
                })
                if (i == -1) {
                    $scope.data.push(data);
                } else {
                    $scope.data[i] = data
                }
            });
        };
        
        $gadgets.getDevices($scope.id, function(locations, directions, targets, method) {
            $scope.sources = {};
            angular.forEach(directions, function(value, key) {
                if (value == "input") {
                    $scope.sources[key] = key == $scope.label;
                }
            });
            console.log($scope.sources);
        });
        
        $scope.setSpan = function() {
            $scope.start = getStart();
            $scope.end = getEnd();
            var keys = _.map($scope.data, function(item) {
                return item.key;
            });
            _.each(keys, function(key) {
                $scope.getData(key);
            });
        };

        $scope.getAll = function() {
            _.each($scope.sources, function(val, key) {
                if (val) {
                    $scope.getData(key);
                }
            });
        }
                
        $scope.addSource = function(key) {
            if ($scope.sources[key]) {
                $scope.sources[key] = false;
                $scope.data = _.without($scope.data, _.findWhere($scope.data, {key: key}));
            } else {
                $scope.sources[key] = true;
                $scope.getData(key);
            }
        };
        
        $scope.toggle = function() {
            $mdSidenav('left').toggle();
        }

        $scope.close = function () {
            $mdSidenav('left').close();
        };
        
        $scope.start = getStart();
        $scope.end = getEnd();
        $scope.getData($scope.label);
    }]);

angular.module('quimby.services')
    .service('$stats', ['$http', function ($http) {
        this.getStats = function(id, name, span, summarize, callback) {
            var end = moment().utc().format();
            var start = moment().utc().subtract(span, "hours").format();
            var vals = [];
            var url = "/api/gadgets/" + id + "/sources/" + name;
            $http.get(
                url,
                {
                    params:
                    {
                        start: start,
                        end:end,
                        summarize: summarize
                    }
                }
            ).success(function(data) {
                callback({
                    key: name,
                    values: _.map(data, function (value) {
                        return [new Date(value.x), value.y];
                    })
                })
            })
        }
    }]);
