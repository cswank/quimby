#!/bin/bash

#cd www
#rm -rf dist
#npm install

CHART=Chart.js/Chart.js
CLEANCHART=Chart-js/Chart.js
sed -i 's|'$CHART'|'$CLEANCHART'|g' app/index.html
SCATTER=Chart.Scatter.js/Chart.Scatter.js
CLEANSCATTER=Chart-Scatter-js/Chart.Scatter.js

sed -i 's|'$SCATTER'|'$CLEANSCATTER'|g' app/index.html

mv app/bower_components/Chart.js app/bower_components/Chart-js
mv app/bower_components/Chart.Scatter.js app/bower_components/Chart-Scatter-js

# ./node_modules/.bin/grunt
# cd ..
# rice embed-go
# godep go build
# rm www-dist.rice-box.go
# cd www
# rm -r dist
# ln -s app dist
# cd ..
