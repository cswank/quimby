'use strict';

angular.module('quimby.directives', [])
    .directive('dragEnd', function() {
        return {
            attrs: {
                callback: '&doneSliding'
            },
            link: function(scope, element, attrs) {
                element.on('$md.dragend', function() {
                    scope.doneSliding();
                })
            }
        }
    });
