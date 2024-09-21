example:
	go run ./example/main.go

test:
	go test -race -count=1 --parallel 8 ./... 

bench:
	go run bench/bench.go

get_coverage_pic:
	gopherbadger -md="README.md,coverage.out"
