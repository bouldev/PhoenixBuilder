const fs=require("fs");
let contents=fs.readdirSync("new_pycs");
for(let i of contents) {
	let c=fs.readFileSync(`new_pycs/${i}`).toString();
	if(c.indexOf(process.argv[2])!=-1) {
		console.log(i);
	}
}