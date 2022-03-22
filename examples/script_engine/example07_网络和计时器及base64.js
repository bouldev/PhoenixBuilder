// example05.js
// 本脚本演示了fetch功能

// fetch
engine.message("Start...");
let x = fetch('https://storage.fastbuilder.pro').then((r)=>{
 	r.text().then((d)=>{
		engine.message(r.statusText);
		for (let k in r.headers._headers) {
			engine.message(k + ':', r.headers.get(k));
		}
		engine.message(d);
	});
});

engine.message("Awaiting...");

// setTimeout, clearTimeout, setInterval and clearInterval
setTimeout(()=>{
	engine.message("Timeout 10s");
}, 1000);

//  atob and btoa
let base64encodedString = btoa("raw string");
let recoveredString = atob(base64encodedString);
engine.message(base64encodedString);
engine.message(recoveredString);

// URL and URLSearchParams
// URL.revokeObjectURL();