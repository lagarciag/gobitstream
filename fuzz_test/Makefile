PHONY: fuzz
fuzz:
	go test ./get64BitsFieldFromSlice/ -fuzz=./ -fuzztime 30s

tests:
	go test ./get64BitsFieldFromSlice/...
