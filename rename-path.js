const fs = require('fs');

var isWin = process.platform === "win32";
var isDarwin = process.platform === "darwin";
let goutpath = __dirname + "/go-src/pion_handler.so";
let libpath
if (isWin) {
  goutpath = __dirname + "\\go-src\\pion_handler.so";
  libpath = __dirname + "\\go-src\\pion_handler.dll";
  fs.rename(goutpath, libpath, (err) => {
    if (err) throw err
    console.log("renamed path!")
  });
} else if (isDarwin) {
  goutpath = __dirname + "/go-src/pion_handler.so";
  libpath = __dirname + "/go-src/pion_handler.dylib";
  fs.rename(goutpath, libpath, (err) => {
    if (err) throw err;
    console.log("renamed path!");
  });
} else { 
  goutpath = __dirname + "/go-src/pion_handler.so";
  libpath = __dirname + "/go-src/pion_handler.so";
}
