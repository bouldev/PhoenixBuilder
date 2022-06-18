// +build is_tweak

package commands
// callee free plz

// void phoenixbuilder_show_title(char *message);
import "C"

func (sender *CommandSender) Title(message string) error {
	C.phoenixbuilder_show_title(C.CString(message))
	return nil
}