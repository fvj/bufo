#include "adapter_abi.h"
#include "tomlc17.h"
#include <arpa/inet.h>
#include <fcntl.h>
#include <limits.h>
#include <netdb.h>
#include <stdio.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

int main(int argc, char *argv[]) {
  if (argc != 2) {
    fprintf(stderr, "usage: %s <file>\n", argv[0]);
    return 1;
  }

  const char *file_path = argv[1];
  toml_result_t result = toml_parse_file_ex(file_path);

  if (!result.ok) {
    fprintf(stderr, "failure to parse toml: %s\n", result.errmsg);
  }

  toml_datum_t adapters = toml_seek(result.toptab, "adapters.enabled");

  if (adapters.type != TOML_ARRAY) {
    fprintf(stderr, "missing or invalid 'adapters' property in config!\n");
    return BUFO_GENERIC_FAILURE;
  }

  for (unsigned int i = 0; i < adapters.u.arr.size; i++) {
    toml_datum_t elem = adapters.u.arr.elem[i];

    if (elem.type != TOML_STRING) {
      fprintf(stderr, "invalid entry in 'adapters': wrong type (%d).\n",
              elem.type);
      return BUFO_GENERIC_FAILURE;
    }

    printf("registering new adapter: %s\n", elem.u.s);
  }

  return 0;
}
