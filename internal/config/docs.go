package config

import "fmt"

type DocsConfig struct {
	Host string
	Port string
}

func (d *DocsConfig) GetAddr() string {
	return fmt.Sprintf("http://%v:%v/swagger/doc.json", d.Host, d.Port)
}
