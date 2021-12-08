package config

type Config struct {
	Agents     []Agent     `mapstructure:"agents"`
	Namespaces []Namespace `mapstructure:"namespaces"`
}

// type Agents struct {
// 	Agent map[string][]Agent `mapstructure:"agents"`
// 	Name  string             `mapstructure:"name"`
// }

type Agent struct {
	Name     string `mapstructure:"name"`
	Fqdn     string `mapstructure:"fqdn"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// type Namespaces struct {
// 	namespace map[string][]Agent `mapstructure:"namespaces"`
// 	Agent     Agent              `mapstructure:"agent"`
// }

type Namespace struct {
	Name      string `mapstructure:"name"`
	AgentName string `mapstructure:"agent"`
	Password  string `mapstructure:"password"`
}
