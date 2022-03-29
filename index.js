const ffi = require('ffi-napi');
//const fs = require('fs');
//const ref = require('ref-napi');
//const Struct = require('ref-struct-di')(ref);

var isWin = process.platform === "win32";
var isDarwin = process.platform === "darwin";

/* adapt ffi code
 * 
const file = join(dirname(fileURLToPath(import.meta.url)), "../dist/regodit.dll")
             .replace('app.asar', 'app.asar.unpacked'); //electron asar friendly
*/
let libpath

if (isWin) {
	libpath = __dirname + "\\go-src\\pion_handler.dll";
} else if (isDarwin) {
	libpath = __dirname + "/go-src/pion_handler.dylib";
} else {
	libpath = __dirname + "/go-src/pion_handler.so";
}

var pionjs = ffi.Library(libpath, {
  SpawnConnection: ["string", ["string"]],
  SetRemoteDescription: ["string", ["bool"]],
  AddIceCandidate: ["string", ["bool"]]
});



module.exports = pionjs;
