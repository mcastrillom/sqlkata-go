module github.com/mcastrillom/sqlkata-go/execution

go 1.22

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/jmoiron/sqlx v1.4.0
	github.com/mcastrillom/sqlkata-go v0.1.1
)

// Local clones only. Consumers ignore replace and resolve v0.1.1 from the module proxy.
replace github.com/mcastrillom/sqlkata-go => ../
