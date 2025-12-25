module github.com/toutaio/toutago-datamapper/examples/mysql

go 1.21.5

require (
	github.com/toutaio/toutago-datamapper v0.1.0
	github.com/toutaio/toutago-datamapper-mysql v0.0.0-00010101000000-000000000000
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/toutaio/toutago-datamapper => ../..

replace github.com/toutaio/toutago-datamapper-mysql => /home/nestor/Proyects/toutago-datamapper-mysql
