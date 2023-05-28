PHONY: fuzz
fuzz:
	go test ./get64BitsFieldFromSlice/ -fuzz=./ -fuzztime 30s
	go test ./readwrite/ -fuzz=./ -fuzztime 30s

tests:
	go test ./get64BitsFieldFromSlice/...
	go test ./readwrite/...
