module example.com/checkout

// Requirements merged into one block and reordered; the versions are the
// same as checkout-v1.
go 1.21

require (
	example.com/payclient v1.4.0
	example.com/httpkit v1.2.0
	example.com/textutil v1.0.0 // indirect
	example.com/telemetry v0.4.0 // indirect
	example.com/retry v1.0.0 // indirect
)
