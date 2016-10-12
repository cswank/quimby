THESE ARE JUST NOTES UNTIL I MAKE A REAL README

Install
=======

You can build Quimby so that the angular front end code is embedded in
the binary.  Do this by calling

    $ ./bin/build.sh

from the root of quimby.

DONT FORGET TO ADD QUIMBY_USER TO DB WITH WRITE PERMISSION

Secure Websockets on IOS
========================

((make me nice, seriously!))
wws doesn't work on iphone unless your certificate organization name matches
your dyn domain and you install the cert on the phone by exporting a .cer file
and emailing it to the phone where it can be installed.
