'use strict';

angular.module('quimby.services')
    .factory('$auth', ['$http', '$location', '$localStorage', function($http, $location, $localStorage) {
        var storage = $localStorage;
        var loggedIn = false;
        var user = false;
        var token = storage.token;
        $http.defaults.headers.common.Authorization = token;
        return {
            getToken: function() {
                return token;
            },
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
                    token = headers().authorization;
                    storage.token = token;
                    $http.defaults.headers.common.Authorization = token;
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
                storage.token = "";
            }
        }
    }]);
