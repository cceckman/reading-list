redo-ifchange \
  ../../app/app.js \
  ../../app/app.js.map

cp ../../app/*.js .
cp ../../app/*.js.map .

sha256sum \
  *.js \
  *.js.map \
| redo-stamp
