
engine.setName("ws server")

// 可以通过 ws://localhost:8888/ws_test 连接
// 即，与例6相同
let server=new ws.Server("0.0.0.0:8888");
let clientIdCounter=0;
server.onconnection=(client)=> {
	engine.message("A new connection established.");
	engine.message(`The client requested path ${client.path}.`);
	let clientId=clientIdCounter;
	clientIdCounter++;
	engine.message(`Let's name it client ${clientId}.`);
	client.onmessage=(msg)=> {
		engine.message(`Message from client ${clientId}: ${msg}`);
		client.send(`ECHO/ ${msg}`);
		engine.message("Echo sent.");
	};
	client.onclose=(msg)=> {
		engine.message(`The connection w/ client ${clientId} is closed.`);
	};
}
engine.message("Server is running at 0.0.0.0:8888");
