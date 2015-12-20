#!/bin/bash

cd www
rm -rf dist
npm install

#hack to deal with grunt suckyness
CHART=Chart.js/Chart.js
CLEANCHART=Chart-js/Chart.js
sed -i 's|'$CHART'|'$CLEANCHART'|g' app/index.html

SCATTER=Chart.Scatter.js/Chart.Scatter.js
CLEANSCATTER=Chart-Scatter-js/Chart.Scatter.js
sed -i 's|'$SCATTER'|'$CLEANSCATTER'|g' app/index.html

ANG=angular-chart.js/dist/angular-chart.js
CLEANANG=angular-chart-js/dist/angular-chart.js
sed -i 's|'$ANG'|'$CLEANANG'|g' app/index.html

mv app/bower_components/Chart.js app/bower_components/Chart-js
mv app/bower_components/Chart.Scatter.js app/bower_components/Chart-Scatter-js
mv app/bower_components/angular-chart.js app/bower_components/angular-chart-js

./node_modules/.bin/grunt

mv app/bower_components/Chart-js app/bower_components/Chart.js 
mv app/bower_components/Chart-Scatter-js app/bower_components/Chart.Scatter.js
mv app/bower_components/angular-chart-js app/bower_components/angular-chart.js 

sed -i 's|'$CLEANCHART'|'$CHART'|g' app/index.html
sed -i 's|'$CLEANSCATTER'|'$SCATTER'|g' app/index.html
sed -i 's|'$CLEANANG'|'$ANG'|g' app/index.html

cd ..
rice embed-go
godep go build
rm www-dist.rice-box.go
cd www
rm -r dist
ln -s app dist
cd ..
