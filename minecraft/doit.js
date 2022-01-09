const fs=require("fs");

function doit(dir) {
	let contents=fs.readdirSync(dir);
	for(let i of contents) {
		let fn=dir+"/"+i;
		if(fs.lstatSync(fn).isDirectory()) {
			doit(fn);
			continue;
		}
		if(i.indexOf(".go")==-1)continue;
		let f=fs.readFileSync(fn).toString();
		f=f.replace(/github\.com\/sandertv\/gophertunnel/g,"phoenixbuilder");
		fs.writeFileSync(fn,f);
	}
}
doit(".");