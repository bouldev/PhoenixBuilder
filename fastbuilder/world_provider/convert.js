const fs=require("fs");
const rids=require("./runtimeIds.json");

let out=`package world_provider

// Auto generated, DO NOT EDIT

import "phoenixbuilder/dragonfly/server/world"

func InitRuntimeIds() {
`;
let unimplementedCounter=0;
for(let i of rids) {
	if(i===null) {
		out+=`\tworld.RegisterUnimplementedBlock(${unimplementedCounter});\n`;
		unimplementedCounter++;
		continue;
	}
	out+=`\tworld.RegisterBlockState("minecraft:${i[0]}",${i[1]})\n`;
}
out+="}\n\n"
out+="func InitRuntimeIdsWithoutMinecraftPrefix() {\n";
unimplementedCounter=0;
for(let i of rids) {
	if(i===null) {
		out+=`\tworld.RegisterUnimplementedBlock(${unimplementedCounter});\n`;
		unimplementedCounter++;
		continue;
	}
	out+=`\tworld.RegisterBlockState("${i[0]}",${i[1]})\n`;
}
out+="}\n\n"
fs.writeFileSync("runtime_ids.go",out);