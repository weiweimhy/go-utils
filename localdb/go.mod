module github.com/weiweimhy/go-utils/v6/localdb

go 1.24.0

require (
	github.com/weiweimhy/go-utils/v6 v6.0.0
	go.etcd.io/bbolt v1.4.3
)

require golang.org/x/sys v0.40.0 // indirect

replace github.com/weiweimhy/go-utils/v6 => ..
