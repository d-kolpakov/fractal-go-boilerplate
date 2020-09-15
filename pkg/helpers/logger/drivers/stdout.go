package drivers

import (
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/logger"
	"github.com/d-kolpakov/logger/drivers/stdout"
)

type STDOUTDriver struct {
	Base  *stdout.STDOUTDriver
	stats *stats.Stats
}

func (s *STDOUTDriver) Init() error {
	return s.Base.Init()
}

func (s *STDOUTDriver) SetStats(st *stats.Stats) {
	s.stats = st
}

func (s *STDOUTDriver) PutMsg(msg logger.Message) error {
	err := s.Base.PutMsg(msg)

	if s.stats != nil {
		switch msg.MessageType {
		case "ALERT":
			requestID, source := s.getFromTags(msg)
			ltrace := s.getLastTrace(msg)
			s.stats.InsertStat("server.alert."+ltrace, nil, nil, nil, nil, &requestID, &source)
		case "ERROR":
			requestID, source := s.getFromTags(msg)
			ltrace := s.getLastTrace(msg)
			s.stats.InsertStat("server.error."+ltrace, nil, nil, nil, nil, &requestID, &source)
		}
	}

	return err
}

func (s *STDOUTDriver) getFromTags(msg logger.Message) (string, string) {
	var requestID, source string

	if val, ok := msg.Tags["requestID"]; ok {
		requestID = val
	}

	if val, ok := msg.Tags["source"]; ok {
		source = val
	}

	return requestID, source
}

func (s *STDOUTDriver) getLastTrace(msg logger.Message) string {
	var res string
	if msg.Stacktrace != nil && msg.Stacktrace.Frames != nil && len(msg.Stacktrace.Frames) >= 1 {
		f := msg.Stacktrace.Frames[0]
		res = fmt.Sprintf(`%s.%s.%s.%d`, f.AbsPath, f.Module, f.Function, f.Lineno)
	}

	return res
}
