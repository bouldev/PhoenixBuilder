const fs=require("fs");

function doit(dir,isRoot) {
	let readme="# PhoenixBuilder-storage\nThis repository is used to store auto-built binaries for [PhoenixBuilder](https://github.com/LNSSPsd/PhoenixBuilder).\n## Index\n";
	let files=fs.readdirSync(dir);
	if(!isRoot) {
		readme+="[(Parent directory)/](../)  \n";
	}
	for(let i of files) {
		let st=fs.lstatSync(`${dir}/${i}`);
		if(st.isDirectory()) {
			readme+=`[${i}/](${i}/)  \n`;
			doit(`${dir}/${i}`);
		}else{
			readme+=`[${i}](${i})  \n`;
		}
	}
	fs.writeFileSync(`${dir}/README.md`,readme);
}

doit(process.argv[2],true);
