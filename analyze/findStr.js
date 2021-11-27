const fs=require("fs");
let contents=fs.readdirSync("new");
for(let i of contents) {
	let c=fs.readFileSync(`new/${i}`).toString();
	if(c.indexOf("SetloadLoadingTime")!=-1) {
		console.log(i);
	}
}