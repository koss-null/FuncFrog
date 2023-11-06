example:
	go run ./example/main.go

test:
	go test -race --parallel 8 ./... 

get_coverage_pic:
	gopherbadger -md="README.md,coverage.out"
