module github.com/buildpack/pack

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/buildpack/lifecycle v0.0.0-20181025145258-a42887a7d29f
	github.com/buildpack/packs v0.0.0-20180824001031-aa30a412923763df37e83f14a6e4e0fe07e11f25
	github.com/docker/docker v0.7.3-0.20180531152204-71cd53e4a197
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.0-20180209012529-399ea8c73916 // indirect
	github.com/fatih/color v1.7.0
	github.com/golang/mock v1.1.1
	github.com/google/go-cmp v0.2.0
	github.com/google/go-containerregistry v0.0.0-20181023232207-eb57122f1bf9
	github.com/google/uuid v0.0.0-20171129191014-dec09d789f3d
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/pkg/errors v0.8.0
	github.com/prometheus/client_golang v0.9.1 // indirect
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910 // indirect
	github.com/prometheus/common v0.0.0-20181020173914-7e9e6cabbd39 // indirect
	github.com/prometheus/procfs v0.0.0-20181005140218-185b4288413d // indirect
	github.com/sclevine/spec v1.0.0
	github.com/sirupsen/logrus v1.2.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	golang.org/x/net v0.0.0-20181023162649-9b4f9f5ad519 // indirect
	golang.org/x/sys v0.0.0-20181025063200-d989b31c8746 // indirect
)

replace github.com/google/go-containerregistry v0.0.0-20181023232207-eb57122f1bf9 => github.com/dgodd/go-containerregistry v0.0.0-20180912122137-611aad063148a69435dccd3cf8475262c11814f6
