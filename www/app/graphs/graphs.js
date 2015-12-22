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
        
        $scope.id = $routeParams.id;
        $scope.label = $routeParams.location + " " + $routeParams.device;
        
        $scope.getDate = function(){
            return function(d){
                return d3.time.format('%c')(new Date(d));  //uncomment for date format
            }
        }
        
        $scope.getSpan = function(span) {
            $stats.getStats($scope.id, $routeParams.location, $routeParams.device, span, function(data) {
                var transformed = _.map(data, function (value) {
                    return [new Date(value.x), value.y];
                });
                $scope.data = [{
                    key: $scope.label,
                    values: transformed
                }]
            })
        }
        $scope.getSpan($scope.spans[3]);
    }]);

angular.module('quimby.services')
    .service('$stats', ['$http', function ($http) {
        this.getStats = function(id, location, name, span, callback) {
            var end = moment().utc().format();
            var start = moment().utc().subtract(span, "hours").format();
            var url = "/api/gadgets/" + id + "/locations/" + location + "/devices/" + name + "/datapoints"
            console.log(url, start, end);
            $http.get(url).success(function(data) {
                if (data == null) {
                    data = [];
                }
                callback(data);
            })
            // $http.get(
            //     url,
            //     {
            //         params:
            //         {
            //             start: start,
            //             end:end
            //         }
            //     }
            // ).success(function(data) {
            //     if (data == null) {
            //         data = [];
            //     }
            //     callback(data);
            // })
        }
    }]);
