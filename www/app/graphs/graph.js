'use strict';

angular.module('quimby.services')
    .service('$stats', ['$http', function ($http) {
        this.getStats = function(id, location, name, span, callback) {
            var end = moment().utc().format();
            var start = moment().utc().subtract(span, "hours").format();
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
                callback(data);
            })
        }
    }]);

angular.module('quimby.graph', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/graphs/:id', {
            templateUrl: 'graphs/graph.html',
            controller: 'GraphCtrl'
        });
    }])

    .controller('GraphCtrl', ['$scope', '$stats', '$routeParams', function($scope, $stats, $routeParams) {
        $scope.spans = [
            {name: "hour", value: 1},
            {name: "day", value: 24},
            {name: "week", value: 7 * 24},
            {name: "month", value: 7 * 24 * 30}
        ]
        $scope.id = $routeParams.id;
        $scope.label = $routeParams.location + " " + $routeParams.device;
        
        $scope.getSpan = function(span) {
            $stats.getStats($scope.id, $routeParams.location, $routeParams.device, span, function(data) {
                $scope.options = {
                    datasetStroke: false,
                    bezierCurve: false,
                    scaleType: "date"
                };
                $scope.labels = [$scope.label];
                $scope.data = [_.each(data, function (value) {
                    value.x = new Date(value.x);
                })];
                $scope.type = 'Scatter';
            });
        }
        $scope.getSpan($scope.spans[0].value);
    }]);

angular.module('quimby.directives')
    .directive('chartScatter', ["ChartJsFactory", function (ChartJsFactory) {
        return new ChartJsFactory('Scatter');
    }])
