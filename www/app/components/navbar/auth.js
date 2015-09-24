'use strict';

angular.module('quimby.services')
    .factory('$auth', ['$http', '$location', function($http, $location) {
        var loggedIn = false;
        var user = false;
        return {
            getUser: function(callback) {
                if (user) {
                    callback(user);
                }
                else  {
                    $http.get('/api/ping')
                        .success(function (data, status, headers, config) {
                            loggedIn = true;
                            $http.get(headers("Location")).success(function(data) {
                                user = data;
                                callback(user);
                            });
                        })
                        .error(function (data, status, headers, config) {
                            loggedIn = false;
                            user = false;
                        });
                }
            },
            login: function(username, password, callback, errorCallback) {
                var u = {username:username, password: password};
                $http({
                    url: '/api/login',
                    method: "POST",
                    data: JSON.stringify(u),
                    headers: {'Content-Type': 'application/json'}
                }).success(function (data, status, headers, config) {
                    loggedIn = true;
                    $http.get(headers('Location')).success(function(data) {
                        user = data;
                        callback(data);
                    });
                }).error(function (data, status, headers, config) {
                    loggedIn = false;
                    errorCallback();
                });
            },
            logout: function(callback) {
                $http({
                    url: '/api/logout',
                    method: "POST",
                    headers: {'Content-Type': 'application/json'}
                }).success(function (data, status, headers, config) {
                    loggedIn = false;
                    callback();
                }).error(function (data, status, headers, config) {
                    
                });
            }
        }
    }]);
