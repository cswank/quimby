'use strict';

angular.module('quimby.graph', ['ngRoute'])
    .config(['$routeProvider', function($routeProvider) {
        $routeProvider.when('/graphs/:id', {
            templateUrl: 'graphs/graph.html',
            controller: 'GraphCtrl'
        });
    }])

    .controller('GraphCtrl', ['$scope', '$stats', '$routeParams', function($scope, $stats, $routeParams) {
        $scope.id = $routeParams.id;
        $scope.label = $routeParams.location + " " + $routeParams.device;
        $stats.getStats($scope.id, $routeParams.location, $routeParams.device, function(data) {
            $scope.options = {
                datasetStroke: false,
                bezierCurve: false,
                scaleType: "date"
            };
            $scope.labels = [$scope.label];
            $scope.data = [_.each(data, function (value) {
                value.x = new Date(value.x);
            })];
            console.log($scope.data);
            $scope.type = 'Scatter';
        });
    }]);

angular.module('quimby.directives')
    .directive('chartScatter', ["ChartJsFactory", function (ChartJsFactory) {
        return new ChartJsFactory('Scatter');
    }])
