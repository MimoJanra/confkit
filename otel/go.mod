module github.com/MimoJanra/confkit/otel

go 1.25.0

require (
	github.com/MimoJanra/confkit v1.0.0
	go.opentelemetry.io/otel v1.41.0
	go.opentelemetry.io/otel/trace v1.41.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/MimoJanra/confkit v1.0.0 => ../
