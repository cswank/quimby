'use strict';

angular.module('quimby.directives', [])
    .directive('dragEnd', function() {
        return {
            attrs: {
                callback: '&doneDragging'
            },
            link: function(scope, element, attrs) {
                element.on('$md.dragend', function() {
                    console.log(attrs);
                    attrs.doneDragging = true;
                })
            }
        }
    });
