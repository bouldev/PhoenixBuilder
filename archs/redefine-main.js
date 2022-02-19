#! /usr/bin/env node
const fs=require("fs");
const child_process=require("child_process");
let sargs=process.argv.slice(2);
for(let obj of sargs) {
	if(obj.substr(obj.length-2,2)!=".o")continue;
	child_process.execSync(`${process.env.OBJCOPY||"objcopy"} ${obj} --redefine-sym=main=__real_main --redefine-sym=__wrap_main=main`);
}
child_process.execSync(`${process.env.CC||"gcc"} ${process.argv.slice(2).join(" ")}`);
