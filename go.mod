module github.com/honeycombio/agentless-integrations-for-aws

// Go 1.27.3 is the first release with crypto/x509 mitigation for CVE-2026-111943
// (compromised eMudhra intermediate CAs), and it is what honeytail 1.12.0 requires.
// Do not lower this: handlers built with an older toolchain still trust chains
// containing those intermediates.
go 1.27.3

require (
	github.com/aws/aws-lambda-go v1.49.0
	github.com/aws/aws-sdk-go v1.55.7
	github.com/honeycombio/honeytail v1.12.0
	github.com/honeycombio/libhoney-go v1.27.1
	github.com/sirupsen/logrus v1.10.2
	github.com/stretchr/testify v1.12.1
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/facebookgo/limitgroup v0.0.0-20150612190941-6abd8d71ec01 // indirect
	github.com/facebookgo/muster v0.0.0-20150708232844-fd3d7953fd52 // indirect
	github.com/go-sql-driver/mysql v1.10.1 // indirect
	github.com/honeycombio/mysqltools v0.0.1 // indirect
	github.com/honeycombio/sqlparser v0.0.0-20180730202938-aab361df519b // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/kr/logfmt v0.0.0-20210122060352-19f9bcb100e6 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.49.0 // indirect
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
)
