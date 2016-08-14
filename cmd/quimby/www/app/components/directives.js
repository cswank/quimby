'use strict';

angular.module('quimby.directives', [])
    .directive('dragEnd', function() {
        return {
            scope: {
                callback: '&doneDragging'
            },
            link: function(scope, element, attrs) {
                element.on('$md.dragend', function() {
                    console.info('Drag Ended');
                })
            }
        }
    });
