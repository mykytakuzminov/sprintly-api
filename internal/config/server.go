package config

type ServerConfig struct {
	Host string
	Port string
}

func (s *ServerConfig) GetAddr() string {
	return s.Host + ":" + s.Port
}

func (s *ServerConfig) LogFields() []interface{} {
	logFields := []interface{}{
		"server_host", s.Host,
		"server_port", s.Port,
	}

	return logFields
}

func (s *ServerConfig) LogFieldsWithErr(err error) []interface{} {
	logFields := []interface{}{
		"error", err,
	}

	logFields = append(logFields, s.LogFields()...)

	return logFields
}
