const ffi = require('ffi-napi');
const fs = require('fs');
//const ref = require('ref-napi');
//const Struct = require('ref-struct-di')(ref);


/* adapt ffi code
 * 
const file = join(dirname(fileURLToPath(import.meta.url)), "../dist/regodit.dll")
             .replace('app.asar', 'app.asar.unpacked'); //electron asar friendly
*/

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

var pionjs = ffi.Library(libpath, {
  SpawnConnection: ["string", ["string"]],
  SetRemoteDescription: ["string", ["bool"]],
  AddIceCandidate: ["string", ["bool"]]
});



module.exports = pionjs;
