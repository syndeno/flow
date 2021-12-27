package config

type Config struct {
	Port        string       `mapstructure:"port"`
	Nameserver  string       `mapstructure:"nameserver"`
	Nameservers []Nameserver `mapstructure:"nameservers"`
	Brokers     []Broker     `mapstructure:"agents"`
	Namespaces  []Namespace  `mapstructure:"namespaces"`
}

/*
   name: kafka_local
   type: kafka
   servers: kf1.unix.ar:9092
   user: inwx
   password: test
   topic_sufix: fnaa.unix.ar
   topic_prefix: fnaa-1.unix.ar_
*/
type Broker struct {
	Name         string `mapstructure:"name"`
	Type         string `mapstructure:"type"`
	Servers      string `mapstructure:"servers"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Topic_sufix  string `mapstructure:"topic_sufix"`
	Topic_prefix string `mapstructure:"topic_prefix"`
}

/*
	namespaces:
	-
		name: flow.unix.ar
		broker: kafka_local
		ns_private: dns-int
		ns_public: dns-ext
*/
type Namespace struct {
	Name       string `mapstructure:"name"`
	Broker     string `mapstructure:"broker"`
	Ns_private string `mapstructure:"ns_private"`
	Ns_public  string `mapstructure:"ns_public"`
}

/*
nameservers:
  -
      name: dns_int
      host: ns1.unix.ar
      keyfile: "./tsig.key"

*/
type Nameserver struct {
	Name    string `mapstructure:"name"`
	Host    string `mapstructure:"host"`
	Keyfile string `mapstructure:"keyfile"`
}
