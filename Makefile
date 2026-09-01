PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

.PHONY: all build install clean

all: build

build:
	go build -o tally ./cmd/tally

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 tally $(DESTDIR)$(BINDIR)/tally

clean:
	rm -f tally
