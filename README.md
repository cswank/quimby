

Install
=======

You can build Quimby so that the angular front end code is embedded in
the binary.  Do this by calling

    $ ./bin/build.sh

from the root of quimby.  

Secure Websockets on IOS
========================

((make me nice))
wws doesn't work on iphone unless your certificate organization name matches
your dyn domain and you install the cert on the phone by exporting a .cer file
and emailing it to the phone where it can be installed.
