'use strict';

angular.module('quimby.graphs', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/graphs/:id', {
            templateUrl: 'graphs/graphs.html',
            controller: 'GraphsCtrl'
        });
    }])
    .controller('GraphsCtrl', ['$scope', '$routeParams', '$stats', function($scope, $routeParams, $stats) {
        $scope.spans = [
            {name: "hour", value: 1},
            {name: "day", value: 24},
            {name: "week", value: 7 * 24},
            {name: "month", value: 7 * 24 * 30}
        ]
        
        $scope.selected = 3;
        
        $scope.id = $routeParams.id;
        $scope.label = $routeParams.location + " " + $routeParams.device;
        
        $scope.getDate = function(){
            return function(d){
                return d3.time.format('%x %X')(new Date(d));  //uncomment for date format
            }
        }

        $scope.getSelectedStyle = function(index) {
            if (index == $scope.selected) {
                return {color: '#3F51B5'};
            } 
            return {};
        }
        
        $scope.getData = function(i) {
            console.log("index", i);
            $scope.selected = i;
            $stats.getStats($scope.id, $routeParams.location, $routeParams.device, $scope.spans[i], function(data) {
                $scope.data = [{
                    key: $scope.label,
                    values: data
                }]
            })
        }
        $scope.getData($scope.selected);
    }]);

angular.module('quimby.services')
    .service('$stats', ['$http', function ($http) {
        this.getStats = function(id, location, name, span, callback) {
            var end = moment().utc().format();
            var start = moment().utc().subtract(span.value, "hours").format();
            var url = "/api/gadgets/" + id + "/locations/" + location + "/devices/" + name + "/datapoints"
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
                if (data == null) {
                    data = [];
                }
                callback(_.map(data, function (value) {
                    return [new Date(value.x), value.y];
                }));
            })
        }
    }]);
