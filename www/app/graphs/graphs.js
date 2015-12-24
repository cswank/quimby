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
        
        $scope.sources = {}
        $scope.sources[$scope.label] = true;
        
        $scope.getDate = function(){
            return function(d){
                return d3.time.format('%x %X')(new Date(d));  //uncomment for date format
            }
        };

        $scope.getSelectedSourceStyle = function(key) {
            if ($scope.sources[key]) {
                return {color: '#3F51B5'};
            }
            return {};
        };

        $scope.getSelectedStyle = function(name) {
            if (name == $scope.selected) {
                return {color: '#3F51B5'};
            } 
            return {};
        };
        
        $scope.getData = function(name) {
            $scope.selected = name;
            $stats.getStats($scope.id, $scope.sources, $scope.spans[name], function(data) {
                console.log("getData", data);
                $scope.data = data;
            })
        };
        
        $gadgets.getDevices($scope.id, function(locations, directions) {
            $scope.choices = [];
            angular.forEach(directions, function(value, key) {
                if (value == "input") {
                    $scope.choices.push(key);
                }
            })
        });

        $scope.addSource = function(key) {
            if ($scope.sources[key]) {
                delete $scope.sources[key];
            } else {
                $scope.sources[key] = true;
            }
            $scope.getData($scope.selected);
        };
        
        $scope.getData($scope.selected);
    }]);

angular.module('quimby.services')
    .service('$stats', ['$http', function ($http) {
        this.getStats = function(id, names, span, callback) {
            var end = moment().utc().format();
            var start = moment().utc().subtract(span, "hours").format();
            var vals = [];
            _.each(names, function(value, name) {
                var url = "/api/gadgets/" + id + "/sources/" + name;
                console.log("url", url, start, end);
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
                    if (!Array.isArray(data)) {
                        data = [];
                    }
                    vals.push({
                        key: name,
                        values: _.map(data, function (value) {
                            return [new Date(value.x), value.y];
                        })
                    })
                    if (vals.length == _.size(names)) {
                        callback(vals);
                    }
                })
            })
                }
    }]);
