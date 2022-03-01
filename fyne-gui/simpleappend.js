const fs=require("fs");

let mJson=JSON.parse(fs.readFileSync("../build/hashes.json"));
mJson.push(`gui~${fs.readFileSync("version").toString()}`);
fs.writeFileSync("../build/hashes.json",JSON.stringify(mJson,null,"\t"))