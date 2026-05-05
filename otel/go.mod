module github.com/MimoJanra/confkit/otel

go 1.24.0

require (
	github.com/MimoJanra/confkit v0.5.1
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
)

require (
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/MimoJanra/confkit v0.5.1 => ../
