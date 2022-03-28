const ffi = require('ffi-napi');
//const ref = require('ref-napi');
//const Struct = require('ref-struct-di')(ref);


/* adapt ffi code
 * 
const file = join(dirname(fileURLToPath(import.meta.url)), "../dist/regodit.dll")
             .replace('app.asar', 'app.asar.unpacked'); //electron asar friendly
*/

var pionjs = ffi.Library("./go-src/pion_handler.so", {
  SpawnConnection: ["string", ["string"]],
  SetRemoteDescription: ["string", ["string"]]
});

console.log(pionjs.SpawnConnection('[{"urls":["stun:stun.l.google.com:19302"]}]'))

module.SpawnConnection = pionjs.SpawnConnection;
module.SetRemoteDescription = pionjs.SetRemoteDescription;
