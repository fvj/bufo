#include "../../adapter_abi.h"

struct __attribute__((packed)) ICMPEcho {
  uint8_t type;
  uint8_t code;
  uint16_t checksum;
  uint16_t identifier;
  uint16_t sequence_number;
  unsigned char data[1];
};

static uint16_t ones_complement_sum(uint16_t op1, uint16_t op2) {
  if (op2 > UINT16_MAX - op1) {
    return op1 - (UINT16_MAX - op2);
  } else {
    return op1 + op2;
  }
}

static struct ICMPEcho new_icmp_echo_request() {
  struct ICMPEcho echo = {.type = 0x8,
                          .code = 0x0,
                          .checksum = 0x0,
                          .identifier = 0x0,
                          .sequence_number = 0x0,
                          .data = ' '};

  return echo;
}

#define BUFO_ADAPTER_ABI_VERSION 1
static int initialized = 0;

static const struct AdapterMetric metrics[] = {{
    .name = "test",
    .type = METRIC_COUNTER,
    .description = "we sometimes emit this lol",
}};

static int adapter_init(void) {
  _Static_assert(sizeof(struct ICMPEcho) == (8 + sizeof(unsigned char)),
                 "invalid memory layout for ICMPEcho!");

  if (initialized) {
    return BUFO_GENERIC_FAILURE;
  }

  initialized = 1;

  return BUFO_SUCCESS;
}

static int adapter_execute(const struct DeclarationNode *node) {
  if (!initialized) {
    return BUFO_GENERIC_FAILURE;
  }

  return BUFO_SUCCESS;
}

static void adapter_shutdown(void) { initialized = 0; }

static const struct Adapter adapter = {
    .abi_version = BUFO_ADAPTER_ABI_VERSION,
    .metrics = metrics,
    .metric_count = sizeof(metrics) / sizeof(metrics[0]),
    .init = adapter_init,
    .execute = adapter_execute,
    .shutdown = adapter_shutdown,
};

BUFO_API const struct Adapter *get_adapter(void) { return &adapter; }
