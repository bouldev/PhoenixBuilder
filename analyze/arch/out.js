const fs=require("fs");
let out="[]byte{";
let hexs=fs.readFileSync(process.argv[2]);
for(let i of hexs) {
	out+="0x"+i.toString(16);
	out+=",";
}
console.log(out);
