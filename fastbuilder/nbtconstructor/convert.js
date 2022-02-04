const fs=require("fs");
const rids=require("./itemRuntimeIds.json");

let out=`package nbtconstructor

// Auto generated, DO NOT EDIT

import (
	"phoenixbuilder/fastbuilder/types"
)

var ItemMap map[string]types.Item = map[string]types.Item {`;

let idCounter=0;
for(let i of rids) {
	if(i) {
		out+=`\n\t"${i.name}": types.Item {\n\t\tName: "${i.name}",\n\t\tNetworkID: ${idCounter},\n\t\tMaxDamage: ${i.maxDamage},\n\t},`;
	}
	idCounter++;
}
out+=`\n}\n`;

fs.writeFileSync("item_runtime_ids.go",out);