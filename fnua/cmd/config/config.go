/*
Copyright © 2021 Emiliano Spinella emilianofs@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
	Prefix   string `mapstructure:"prefix"`
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
