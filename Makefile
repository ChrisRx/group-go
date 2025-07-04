test:
	@cd tests; go test .

testv:
	@cd tests; go test -v ./... -count 1
