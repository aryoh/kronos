PREFIX=/usr/local
DESTDIR=
GOFLAGS=
BINDIR=${PREFIX}/bin
DATADIR=${PREFIX}/share

PINBA_SRCS = $(wildcard pinba/*.go)
DUMP_SRCS = $(wildcard dump/*.go)

BINARIES = pinba dump
BLDDIR = build

all: $(BINARIES) #$(EXAMPLES)

$(BLDDIR)/%:
	mkdir -p $(dir $@)
	cd $* && go build ${GOFLAGS} -o $(abspath $(BLDDIR)/kronos_$*)

$(BINARIES): %: $(BLDDIR)/%

# Dependencies
$(BLDDIR)/kronos_pinba: $(PINBA_SRCS)
$(BLDDIR)/kronos_dump: $(DUMP_SRCS)

clean:
	rm -fr $(BLDDIR)

# Targets
.PHONY: install clean all
# Programs
.PHONY: $(BINARIES)

install: $(BINARIES) # $(EXAMPLES)
	install -m 755 -d ${DESTDIR}${BINDIR}
	install -m 755 $(BLDDIR)/kronos_pinba ${DESTDIR}${BINDIR}/kronos_pinba
	install -m 755 $(BLDDIR)/kronos_dump ${DESTDIR}${BINDIR}/kronos_dump
