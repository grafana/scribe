package state

import (
	"io"
	"io/fs"
	"os"

	"github.com/sirupsen/logrus"
)

type WriterLogWrapper struct {
	Writer
	Log logrus.FieldLogger
}

func (s *WriterLogWrapper) SetString(arg Argument, val string) error {
	s.Log.Debugf("Setting string in state for '%s' argument '%s'...", arg.Type, arg.Key)
	err := s.Writer.SetString(arg, val)
	if err != nil {
		s.Log.WithError(err).Debugf("Error setting string in state for '%s' argument '%s'", arg.Type, arg.Key)
	}
	return err
}
func (s *WriterLogWrapper) SetInt64(arg Argument, val int64) error {
	s.Log.Debugf("Setting int64 in state for '%s' argument '%s'...", arg.Type, arg.Key)
	err := s.Writer.SetInt64(arg, val)
	if err != nil {
		s.Log.WithError(err).Debugf("Error setting int64 in state for '%s' argument '%s'", arg.Type, arg.Key)
	}
	return err
}
func (s *WriterLogWrapper) SetFloat64(arg Argument, val float64) error {
	s.Log.Debugf("Setting float64 in state for '%s' argument '%s'...", arg.Type, arg.Key)
	err := s.Writer.SetFloat64(arg, val)
	if err != nil {
		s.Log.WithError(err).Debugf("Error setting float64 in state for '%s' argument '%s'", arg.Type, arg.Key)
	}

	return err
}
func (s *WriterLogWrapper) SetFile(arg Argument, val string) error {
	s.Log.Debugf("Setting file in state for '%s' argument '%s'", arg.Type, arg.Key)
	err := s.Writer.SetFile(arg, val)
	if err != nil {
		s.Log.WithError(err).Debugf("Error setting file in state for '%s' argument '%s'", arg.Type, arg.Key)
	}

	return err
}
func (s *WriterLogWrapper) SetFileReader(arg Argument, val io.Reader) (string, error) {
	s.Log.Debugf("Setting file (using io.Reader) in state for '%s' argument '%s'", arg.Type, arg.Key)
	v, err := s.Writer.SetFileReader(arg, val)
	if err != nil {
		s.Log.WithError(err).Debugf("Error setting file (using io.Reader) in state for '%s' argument '%s'", arg.Type, arg.Key)
	}

	return v, err
}
func (s *WriterLogWrapper) SetDirectory(arg Argument, val string) error {
	s.Log.Debugf("Setting directory in state for '%s' argument '%s'", arg.Type, arg.Key)
	err := s.Writer.SetDirectory(arg, val)
	if err != nil {
		s.Log.WithError(err).Debugf("Error setting directory in state for '%s' argument '%s'", arg.Type, arg.Key)
	}

	return err
}

type ReaderLogWrapper struct {
	Reader
	Log logrus.FieldLogger
}

func (s *ReaderLogWrapper) Exists(arg Argument) (bool, error) {
	s.Log.Debugf("Checking state that '%s' argument '%s' exists...", arg.Type, arg.Key)
	v, err := s.Reader.Exists(arg)
	if err != nil {
		s.Log.Debugf("Error getting state for '%s' key '%s'", arg.Type, arg.Key)
	}

	return v, err
}
func (s *ReaderLogWrapper) GetString(arg Argument) (string, error) {
	s.Log.Debugf("Getting string from state for '%s' argument '%s'", arg.Type, arg.Key)
	v, err := s.Reader.GetString(arg)
	if err != nil {
		s.Log.Debugf("Error getting string from state for '%s' key '%s'", arg.Type, arg.Key)
	}

	return v, err
}

func (s *ReaderLogWrapper) GetInt64(arg Argument) (int64, error) {
	s.Log.Debugf("Getting int64 from state for '%s' argument '%s'", arg.Type, arg.Key)
	v, err := s.Reader.GetInt64(arg)
	if err != nil {
		s.Log.Debugf("Error getting int64 from state for '%s' key '%s'", arg.Type, arg.Key)
	}

	return v, err
}
func (s *ReaderLogWrapper) GetFloat64(arg Argument) (float64, error) {
	s.Log.Debugf("Getting float from state for '%s' argument '%s'", arg.Type, arg.Key)
	v, err := s.Reader.GetFloat64(arg)
	if err != nil {
		s.Log.Debugf("Error getting float64 from state for '%s' key '%s'", arg.Type, arg.Key)
	}

	return v, err
}

func (s *ReaderLogWrapper) GetFile(arg Argument) (*os.File, error) {
	s.Log.Debugf("Getting file from state for '%s' argument '%s'", arg.Type, arg.Key)
	v, err := s.Reader.GetFile(arg)
	if err != nil {
		s.Log.Debugf("Error getting file from state for '%s' key '%s'", arg.Type, arg.Key)
	}

	return v, err
}

func (s *ReaderLogWrapper) GetDirectory(arg Argument) (fs.FS, error) {
	s.Log.Debugf("Getting directory (fs.FS) from state for '%s' argument '%s'", arg.Type, arg.Key)
	v, err := s.Reader.GetDirectory(arg)
	if err != nil {
		s.Log.Debugf("Error getting directory from state for '%s' key '%s'", arg.Type, arg.Key)
	}

	return v, err
}

func (s *ReaderLogWrapper) GetDirectoryString(arg Argument) (string, error) {
	s.Log.Debugf("Getting directory (string) from state for '%s' argument '%s'", arg.Type, arg.Key)
	v, err := s.Reader.GetDirectoryString(arg)
	if err != nil {
		s.Log.Debugf("Error getting int64 state for '%s' key '%s'", arg.Type, arg.Key)
	}

	return v, err
}

type HandlerLogWrapper struct {
	*ReaderLogWrapper
	*WriterLogWrapper
}

func WriterWithLogs(log logrus.FieldLogger, state Writer) *WriterLogWrapper {
	return &WriterLogWrapper{
		Writer: state,
		Log:    log,
	}
}

func ReaderWithLogs(log logrus.FieldLogger, state Reader) *ReaderLogWrapper {
	return &ReaderLogWrapper{
		Reader: state,
		Log:    log,
	}
}

func HandlerWithLogs(log logrus.FieldLogger, state Handler) *HandlerLogWrapper {
	return &HandlerLogWrapper{
		ReaderLogWrapper: &ReaderLogWrapper{
			Log:    log,
			Reader: state,
		},
		WriterLogWrapper: &WriterLogWrapper{
			Log:    log,
			Writer: state,
		},
	}
}
