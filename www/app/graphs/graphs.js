'use strict';

angular.module('quimby.graphs', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/graphs/:id', {
            templateUrl: 'graphs/graphs.html',
            controller: 'GraphsCtrl'
        });
    }])
    .controller('GraphsCtrl', ['$scope', '$routeParams', '$stats', '$gadgets', function($scope, $routeParams, $stats, $gadgets) {
        $scope.spans = {
            hour: 1,
            day: 24,
            week: 7 * 24,
            month: 7 * 24 * 30
        };
        $scope.spanLabels = ["hour", "day", "week", "month"];
        
        $scope.selected = "hour";
        
        $scope.id = $routeParams.id;
        $scope.label = $routeParams.location + " " + $routeParams.device;

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
