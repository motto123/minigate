module com.example.client

go 1.17

replace (
	com.minigame.component => ../../component
	//com.minigame.server/gate/codec => ../../server/gate/codec
	com.minigame.proto => ../../proto
	com.minigame.server.gate => ../../server/gate
	com.minigame.utils => ../../utils
)

require (
	com.minigame.component v0.0.0-00010101000000-000000000000
	com.minigame.proto v0.0.0-00010101000000-000000000000
	github.com/bwmarrin/snowflake v0.3.0
	github.com/pkg/errors v0.9.1
	google.golang.org/protobuf v1.28.0
)

require (
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.10.1 // indirect
	github.com/streadway/amqp v1.0.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
