# build/adapter.mk

ifndef ADAPTER_NAME
$(error ADAPTER_NAME not set)
endif
ifndef ADAPTER_SRCS
$(error ADAPTER_SRCS not set)
endif
ifndef BIN_DIR
$(error BIN_DIR not set)
endif

$(foreach s,$(ADAPTER_SRCS),\
  $(if $(wildcard $(s)),,\
    $(error Source file not found: $(s))))


OBJ_DIR := $(BIN_DIR)/obj/$(ADAPTER_NAME)
OBJS := $(ADAPTER_SRCS:%.c=$(OBJ_DIR)/%.o)

.PHONY: all clean

all: $(BIN_DIR)/$(ADAPTER_NAME).so

$(OBJ_DIR)/%.o: %.c
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) $(CPPFLAGS) $(PICFLAGS) $(DEPFLAGS) -c $< -o $@

$(BIN_DIR)/$(ADAPTER_NAME).so: $(OBJS)
	@mkdir -p $(dir $@)
	$(CC) $(LDFLAGS) $(SO_LDFLAGS) -o $@ $^

clean:
	rm -rf $(OBJ_DIR)
