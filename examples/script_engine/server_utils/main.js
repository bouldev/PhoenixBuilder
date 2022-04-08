// Server utils script
printf("Server utils script\n");

let config={}
//fs.requireContainer("com.lns.serverutils");
if(!fs.exists("config.json")) {
	printf("Initializing container\n");
	fs.writeFile("config.json",JSON.stringify(config, null, "\t"));
	printf("Saved default configurations to %sconfig.json .\n",fs.containerPath);
}else{
	config=JSON.parse(fs.readFile("config.json"))
	console.log("Loaded configurations");
}

engine.waitConnectionSync();
let playerList={};
game.subscribePacket("IDPlayerList", (packet)=>{
	if(pk.ActionType==0) {
		for(let i of packet.Entries) {
			playerList[i.Username]={};
		}
	}
});
game.listenChat((user,msg)=>{
	if(user.length==0)return;
	let splittedMessage=msg.split(" ");
	if(splittedMessage[0]=="")return;
	for(let i=splittedMessage.length-1;i>=0;i--) {
		if(splittedMessage[i]=="") {
			splittedMessage.splice(i,1);
		}
	}
	//if(splittedMessage[0]==".tpa") {
	//	if(splittedMessage[1]=="disable") {
	//		game.oneShotCommand(sprintf("tellraw %s %s",user,JSON.stringify({
});