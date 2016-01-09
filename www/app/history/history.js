'use strict';

angular.module('quimby.history', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/gadgets/:id/history', {
            templateUrl: 'history/history.html',
            controller: 'HistoryCtrl'
        });
    }])
    .controller('HistoryCtrl', ['$scope', '$routeParams', '$stats', '$gadgets', '$window', function($scope, $routeParams, $stats, $gadgets, $window) {
        $scope.spans = {
            hour: 1,
            day: 24,
            week: 7 * 24,
            month: 7 * 24 * 30
        };
        $scope.spanLabels = ["hour", "day", "week", "month"];
        
        $scope.selected = "hour";
        $scope.label = $routeParams.location + " " + $routeParams.device;
        $scope.id = $routeParams.id;

        $gadgets.getGadget($scope.id, function(data) {
            $scope.gadget = data;
        });

        function getEnd() {
            return encodeURIComponent(moment().utc().format());
        }

        function getStart() {
            return encodeURIComponent(moment().utc().subtract($scope.spans[$scope.selected], "hours").format());
        }

        $scope.data = [];
        
        $scope.getDate = function(){
            return function(d){
                return d3.time.format('%x %X')(new Date(d));  //uncomment for date format
            }
        };

        function isSelected (key) {
            return _.findIndex($scope.data, function(item) {
                return item.key == key;
            }) > -1;

        };

        $scope.getHeight = function() {
            return 2 * $window.innerHeight / 3;
        }

        $scope.getSourceStyle = function(key) {
            if (isSelected(key)) {
                return {color: '#3F51B5'};
            }
            return {};
        }

        $scope.getSpanStyle = function(name) {
            if (name == $scope.selected) {
                return {color: '#3F51B5'};
            }
            return {};
        };
        
        $scope.getData = function(key) {
            $stats.getStats($scope.id, key, $scope.spans[$scope.selected], function(data) {
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
        
        $gadgets.getDevices($scope.id, function(locations, directions) {
            $scope.choices = [];
            angular.forEach(directions, function(value, key) {
                if (value == "input") {
                    $scope.choices.push(key);
                }
            });
        });
        
        $scope.setSpan = function(name) {
            $scope.selected = name;
            $scope.start = getStart();
            $scope.end = getEnd();
            var keys = _.map($scope.data, function(item) {
                return item.key;
            });
            _.each(keys, function(key) {
                $scope.getData(key);
            });
        };
        
        $scope.addSource = function(key) {
            if (isSelected(key)) {
                $scope.data = _.without($scope.data, _.findWhere($scope.data, {key: key}));
            } else {
                $scope.getData(key);
            }
        };
        
        $scope.start = getStart();
        $scope.end = getEnd();
        $scope.getData($scope.label);
    }]);

angular.module('quimby.services')
    .service('$stats', ['$http', function ($http) {
        this.getStats = function(id, name, span, callback) {
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
                        end:end
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
