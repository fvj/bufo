BINARIES := bufo
ADAPTERS := icmp

OPTIMIZATION_LEVEL ?= -O2
CC ?= cc

CPPFLAGS := \
	-D_FORTIFY_SOURCE=2 \
	-D_DEFAULT_SOURCE \
	-D_POSIX_C_SOURCE=200809L

CFLAGS := \
	-fcolor-diagnostics \
	-fvisibility=hidden \
	-std=c23 \
	$(OPTIMIZATION_LEVEL) \
	-g \
	-fno-omit-frame-pointer \
	-fstack-protector-strong \
	-fstack-clash-protection \
	-Wall -Wextra -Wpedantic

LDFLAGS := \
	-Wl,-z,relro \
	-Wl,-z,now

DEPFLAGS := -MMD -MP
PICFLAGS := -fPIC
SO_LDFLAGS := -shared -Wl,-soname,$(@F)

.PHONY: all clean adapters $(ADAPTERS)

all: $(BINARIES) adapters

# ---- core binary ----

bufo_SRCS := src/bufo.c src/tomlc17.c
bufo_OBJS := $(bufo_SRCS:src/%.c=bin/%.o)

bin/%.o: src/%.c
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) $(CPPFLAGS) $(DEPFLAGS) -c $< -o $@

bin/bufo: $(bufo_OBJS)
	@mkdir -p bin
	$(CC) $(LDFLAGS) -o $@ $^

%: bin/%
	@:

# ---- adapters ----

adapters: $(ADAPTERS)

$(ADAPTERS):
	$(MAKE) -C src/adapters/$@ \
		BIN_DIR=$(CURDIR)/bin/adapters \
		CC="$(CC)" \
		CFLAGS="$(CFLAGS)" \
		CPPFLAGS="$(CPPFLAGS)" \
		LDFLAGS="$(LDFLAGS)" \
		PICFLAGS="$(PICFLAGS)" \
		SO_LDFLAGS="$(SO_LDFLAGS)" \
		DEPFLAGS="$(DEPFLAGS)"

clean:
	rm -rf bin
	for a in $(ADAPTERS); do \
		$(MAKE) -C src/adapters/$$a clean || true; \
	done

-include $(shell find bin -name '*.d' 2>/dev/null)
