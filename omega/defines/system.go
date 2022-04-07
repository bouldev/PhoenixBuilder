package defines

type System interface {
	Stop() error
	Activate(Adaptor)
}
