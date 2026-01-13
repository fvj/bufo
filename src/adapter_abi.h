/* Defines shared structures and definitions for the bufo plugins */

#ifndef BUFO_ADAPTER_ABI_H
#define BUFO_ADAPTER_ABI_H
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define BUFO_SUCCESS ((unsigned char)0)
#define BUFO_GENERIC_FAILURE ((unsigned char)1 << 8)

#define BUFO_API __attribute__((visibility("default")))

/* Scenario Handlers
 *
 * Helpers for the adapter to handle the passed in declaration.
 * The adapter itself should not know what way the data is represented
 * (TOML, YAML, etc.) and should not have to link with any parsing library.
 *
 * The passed Node effectively acts like a handle to the scenario definition.
 */
struct DeclarationNode;

enum DeclarationGetStatus {
  DECL_OK = BUFO_SUCCESS,
  DECL_MISSING_KEY,
  DECL_TYPE_MISMATCH,
  DECL_INVALID_VALUE
};

enum DeclarationGetStatus
declaration_get_string(const struct DeclarationNode *node, const char *key,
                       const char **out);

enum DeclarationGetStatus
declaration_get_uint32(const struct DeclarationNode *node, const char *key,
                       uint32_t *out);

enum DeclarationGetStatus
declaration_get_uint64(const struct DeclarationNode *node, const char *key,
                       uint64_t *out);

enum DeclarationGetStatus
declaration_get_int32(const struct DeclarationNode *node, const char *key,
                      int32_t *out);

enum DeclarationGetStatus
declaration_get_int64(const struct DeclarationNode *node, const char *key,
                      int64_t *out);

enum DeclarationGetStatus
declaration_get_float(const struct DeclarationNode *node, const char *key,
                      float *out);

enum DeclarationGetStatus
declaration_get_double(const struct DeclarationNode *node, const char *key,
                       double *out);

enum MetricType { METRIC_COUNTER, METRIC_GAUGE, METRIC_SUM, METRIC_AVG };

struct AdapterMetric {
  const char *name;
  enum MetricType type;
  const char *description;
};

struct Adapter {
  uint32_t abi_version;

  /* Static metadata */
  const struct AdapterMetric *metrics;
  size_t metric_count;

  /* Lifecycle methods
   * init and shutdown may be NULL if no init or shutdown, respectively,
   * are needed. */
  int (*init)(void);
  /* return value indicates success. non-zero indicates failure with error */
  int (*execute)(const struct DeclarationNode *step);
  void (*shutdown)(void);
};

/* to be implemented by every adapter */
const struct Adapter *get_adapter(void);

#ifdef __cplusplus
}
#endif

#endif
