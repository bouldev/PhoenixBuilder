const fs=require("fs");
const crypto=require("crypto");

function sha256(data){
	return crypto.createHash("sha256").update(data).digest("hex");
}

let hashes={};
let files=fs.readdirSync("build");
for(let i of files){
	if(i.indexOf(".a")!=-1||i.indexOf(".h")!=-1)continue;
	hashes[i]=sha256(fs.readFileSync(`build/${i}`));
}
fs.writeFileSync("build/hashes.json",JSON.stringify(hashes,null,"\t"));