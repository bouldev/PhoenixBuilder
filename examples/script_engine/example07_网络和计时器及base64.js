// example05.js
// 本脚本演示了fetch功能

// fetch
engine.message("Start...")
var x = fetch('https://storage.fastbuilder.pro').then(function (r) {
    r.text().then(function (d) {
        engine.message(r.statusText)
        for (var k in r.headers._headers) {
            engine.message(k + ':', r.headers.get(k))
        }
        engine.message(d)
    });
});

engine.message("Awaiting...")

// setTimeout, clearTimeout, setInterval and clearInterval
setTimeout(function () {
    engine.message("Timeout 10s")
}, 1000)

//  atob and btoa
base64encodedString = btoa("raw string")
recoveredString = atob(base64encodedString)
engine.message(base64encodedString)
engine.message(recoveredString)

// URL and URLSearchParams
// URL.revokeObjectURL()