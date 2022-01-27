package world_provider

type StubLogger struct {}

func (*StubLogger) Debugf(format string, v ...interface{}) {}
func (*StubLogger) Infof(format string, v ...interface{}) {}
func (*StubLogger) Errorf(format string, v ...interface{}) {}
func (*StubLogger) Fatalf(format string, v ...interface{}) {}
