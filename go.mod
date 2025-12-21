module github.com/ksrzmv/xch

go 1.25.1

require (
	github.com/google/uuid v1.6.0
	github.com/ksrzmv/krypto/krypto v0.0.0-20251221090416-db42f198fc85
	github.com/lib/pq v1.10.9
)

replace github.com/ksrzmv/xch/pkg/message => ./pkg/message

replace github.com/ksrzmv/xch/pkg/misc => ./pkg/misc
