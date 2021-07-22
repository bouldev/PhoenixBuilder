const fs=require("fs");
let builder=fs.readFileSync(process.argv[2]);
let index;
function killKeywordAto(keyword){
	while((index=builder.indexOf(keyword))!=-1){
		while(builder[index]!=0){
			builder[index]=119;
			index++;
		}
	}
}
function killKeyword(keyword,wordPool){
	while((index=builder.indexOf(keyword))!=-1){
		for(let i=0;i<keyword.length;i++){
			builder[index+i]=wordPool?wordPool.charCodeAt(i):42;
		}
	}
}
killKeywordAto("cv4");
killKeywordAto("runtime/debug");
killKeywordAto("main.main");
killKeyword(".go",".rs");
killKeyword("Hash","wwww");
killKeyword("hash","WWWW");
killKeyword("goroutine","rustruntm");
fs.writeFileSync(process.argv[2],builder);