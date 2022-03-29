const ffi = require('ffi-napi');
const fs = require('fs');
//const ref = require('ref-napi');
//const Struct = require('ref-struct-di')(ref);

var isWin = process.platform === "win32";
var isDarwin = process.platform === "darwin";
var isElectron = (typeof process !== 'undefined' && typeof process.versions === 'object' && !!process.versions.electron)

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

if (isElectron) {
  let possibleLibPath = libpath.replace('app.asar', 'app.asar.unpacked')
  if (fs.existsSync(possibleLibPath)) {
    libpath = possibleLibPath
  }
}

var pionjs = ffi.Library(libpath, {
  SpawnConnection: ["string", ["string"]],
  SetRemoteDescription: ["bool", ["string"]],
  AddIceCandidate: ["bool", ["string"]],
  CloseConnection: ["bool", []]
});




module.exports = pionjs;
