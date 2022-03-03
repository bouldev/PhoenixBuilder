// example05.js
// 本脚本演示了fetch功能

var x = fetch('https://storage.fastbuilder.pro').then(function(r) {
    r.text().then(function(d) {
        FB_Println(r.statusText)
        for (var k in r.headers._headers) {
            FB_Println(k + ':', r.headers.get(k))
        }
        FB_Println(d)
    });
});

FB_Println("Awaiting...")

