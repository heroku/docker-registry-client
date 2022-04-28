module github.com/heroku/docker-registry-client

go 1.18

require (
	github.com/docker/distribution v0.0.0-20171011171712-7484e51bf6af
	github.com/golangci/golangci-lint v1.17.2-0.20190909185456-6163a8a79084
	github.com/opencontainers/go-digest v1.0.0-rc1
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/OpenPeeDeeP/depguard v1.0.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/fatih/color v1.6.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/go-critic/go-critic v0.3.5-0.20190526074819-1df300866540 // indirect
	github.com/go-lintpack/lintpack v0.5.2 // indirect
	github.com/go-toolsmith/astcast v1.0.0 // indirect
	github.com/go-toolsmith/astcopy v1.0.0 // indirect
	github.com/go-toolsmith/astequal v1.0.0 // indirect
	github.com/go-toolsmith/astfmt v1.0.0 // indirect
	github.com/go-toolsmith/astp v1.0.0 // indirect
	github.com/go-toolsmith/strparse v1.0.0 // indirect
	github.com/go-toolsmith/typep v1.0.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang/mock v1.0.0 // indirect
	github.com/golangci/check v0.0.0-20180506172741-cfe4005ccda2 // indirect
	github.com/golangci/dupl v0.0.0-20180902072040-3e9179ac440a // indirect
	github.com/golangci/errcheck v0.0.0-20181223084120-ef45e06d44b6 // indirect
	github.com/golangci/go-misc v0.0.0-20180628070357-927a3d87b613 // indirect
	github.com/golangci/go-tools v0.0.0-20190318055746-e32c54105b7c // indirect
	github.com/golangci/goconst v0.0.0-20180610141641-041c5f2b40f3 // indirect
	github.com/golangci/gocyclo v0.0.0-20180528134321-2becd97e67ee // indirect
	github.com/golangci/gofmt v0.0.0-20181222123516-0b8337e80d98 // indirect
	github.com/golangci/gosec v0.0.0-20190211064107-66fb7fc33547 // indirect
	github.com/golangci/ineffassign v0.0.0-20190609212857-42439a7714cc // indirect
	github.com/golangci/lint-1 v0.0.0-20190420132249-ee948d087217 // indirect
	github.com/golangci/maligned v0.0.0-20180506175553-b1d89398deca // indirect
	github.com/golangci/misspell v0.0.0-20180809174111-950f5d19e770 // indirect
	github.com/golangci/prealloc v0.0.0-20180630174525-215b22d4de21 // indirect
	github.com/golangci/revgrep v0.0.0-20180526074752-d9c87f5ffaf0 // indirect
	github.com/golangci/unconvert v0.0.0-20180507085042-28b1c447d1f4 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gostaticanalysis/analysisutil v0.0.0-20190318220348-4088753ea4d3 // indirect
	github.com/hashicorp/hcl v0.0.0-20180404174102-ef8a98b0bbce // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kisielk/gotool v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/magiconair/properties v1.7.6 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.3 // indirect
	github.com/mitchellh/go-homedir v1.0.0 // indirect
	github.com/mitchellh/mapstructure v0.0.0-20180220230111-00c29f56e238 // indirect
	github.com/nbutton23/zxcvbn-go v0.0.0-20171102151520-eafdab6b0663 // indirect
	github.com/pelletier/go-toml v1.1.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/sourcegraph/go-diff v0.5.1 // indirect
	github.com/spf13/afero v1.1.0 // indirect
	github.com/spf13/cast v1.2.0 // indirect
	github.com/spf13/cobra v0.0.2 // indirect
	github.com/spf13/jwalterweatherman v0.0.0-20180109140146-7c0cea34c8ec // indirect
	github.com/spf13/pflag v1.0.1 // indirect
	github.com/spf13/viper v1.0.2 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/timakin/bodyclose v0.0.0-20190721030226-87058b9bfcec // indirect
	github.com/ultraware/funlen v0.0.1 // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/text v0.3.0 // indirect
	golang.org/x/tools v0.0.0-20190521203540-521d6ed310dd // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	mvdan.cc/interfacer v0.0.0-20180901003855-c20040233aed // indirect
	mvdan.cc/lint v0.0.0-20170908181259-adc824a0674b // indirect
	mvdan.cc/unparam v0.0.0-20190209190245-fbb59629db34 // indirect
	sourcegraph.com/sqs/pbtypes v0.0.0-20180604144634-d3ebe8f20ae4 // indirect
)

// From/For golangci-lint. Can be removed once v1.17.2 (or newer) is released
replace (
	// https://github.com/ultraware/funlen/pull/1
	github.com/ultraware/funlen => github.com/golangci/funlen v0.0.0-20190909161642-5e59b9546114
	// https://github.com/golang/tools/pull/139
	golang.org/x/tools => github.com/golangci/tools v0.0.0-20190909104219-979bdb7f8cc8
)
