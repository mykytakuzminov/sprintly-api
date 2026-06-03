package config

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (r *RedisConfig) GetAddr() string {
	return r.Host + ":" + r.Port
}

func (r *RedisConfig) LogFields() []interface{} {
	logFields := []interface{}{
		"redis_host", r.Host,
		"redis_port", r.Port,
		"redis_db", r.DB,
	}

	return logFields
}

func (r *RedisConfig) LogFieldsWithErr(err error) []interface{} {
	logFields := []interface{}{
		"error", err,
	}

	logFields = append(logFields, r.LogFields()...)

	return logFields
}
