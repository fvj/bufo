# no builtin (implicit rules) pls
# everything must be explicit in this house!!!
.SUFFIXES:

# don't delete my builds pls
# (keep intermediates)
.SECONDARY:

CC ?= cc

# modern C please, incl. glibc extensions
CPPFLAGS ?= \
	-D_FORTIFY_SOURCE=2 \
	-D_DEFAULT_SOURCE \
	-D_POSIX_C_SOURCE=200809L

# protect me from what I want
# latest C dialect, safe optimizations only, debug info (strip if you want to),
# stack traces, reasonably aggressive warnings and errors
CFLAGS ?= \
	-std=c23 \
	-O2 \
	-g \
	-fno-omit-frame-pointer \
	-fstack-protector-strong \
	-fstack-clash-protection \
	-Wall \
	-Wextra \
	-Wpedantic \
	-Wshadow \
	-Wcast-qual \
	-Wcast-align \
	-Wpointer-arith \
	-Wstrict-prototypes \
	-Wmissing-prototypes \
	-Wmissing-declarations \
	-Wformat=2 \
	-Wundef \
	-Wwrite-strings \
	-Wno-unused-parameter

LDFLAGS ?= \
	-Wl,-z,relro \
	-Wl,-z,now

DEPFLAGS = -MMD -MP

.PHONY: bindir
bindir:
	@mkdir -p bin

bin/%.o: src/%.c | bindir
	$(CC) $(CFLAGS) $(CPPFLAGS) $(DEPFLAGS) -c $< -o $@

bin/%: bin/%.o | bindir
	$(CC) $(LDFLAGS) -o $@ $<

%: bin/%
	@:

-include bin/*.d
