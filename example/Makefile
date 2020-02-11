ARTIFUCT := ./main
GO := `which go`
GOSRC := \
	./main.go \
	../logger.go \
	../rotator/rotator.go \

build_and_run: $(ARTIFUCT)
	$(ARTIFUCT)

$(ARTIFUCT): $(GOSRC)
	$(GO) build -o $(ARTIFUCT)

test: $(ARTIFUCT)
	$(GO) test ./...

clean:
	rm -f $(ARTIFUCT)
	rm ./*.log
