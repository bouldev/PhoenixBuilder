// +build is_tweak

package commands

import "fmt"

/*
void phoenixbuilder_output(char *content);
void phoenixbuilder_worldchat_output(char *formatted_string, char *sender, char *content);

*/
import "C"

func (sender *CommandSender) Output(content string) error {
	//bridge_fmt.Printf("%s\n", content)
	//if(!args.InGameResponse()) {
	//	return nil
	//}
	C.phoenixbuilder_output(C.CString(content))
	return nil
}

func (cmd_sender *CommandSender) WorldChatOutput(sender string, content string) error {
	//bridge_fmt.Printf("W <%s> %s\n", sender, content)
	str:=fmt.Sprintf("§eW §r<%s> %s",sender,content)
	C.phoenixbuilder_worldchat_output(C.CString(str), C.CString(sender), C.CString(content))
	return nil
}
