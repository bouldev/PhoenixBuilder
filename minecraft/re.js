const fs=require("fs");

function doit(path) {
	let content=fs.readdirSync(path);
	for(let fn of content) {
		let p=path+"/"+fn;
		if(fs.lstatSync(p).isDirectory()) {
			doit(p);
		}else{
			if(p.indexOf(".go")==-1)continue;
			fs.writeFileSync(p, fs.readFileSync(p).toString().replace(/github\.com\/sandertv\/gophertunnel\/minecraft/g, "phoenixbuilder/minecraft"));
		}
	}
}

doit(".");