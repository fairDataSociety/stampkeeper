package mock

import "log"

type Logging struct{}

func (Logging) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (Logging) Info(args ...interface{}) {
	log.Println(args...)
}

func (Logging) Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (Logging) Debug(args ...interface{}) {
	log.Println(args...)
}

func (Logging) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (Logging) Error(args ...interface{}) {
	log.Println(args...)
}

func (Logging) Warningf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (Logging) Warning(args ...interface{}) {
	log.Println(args...)
}
