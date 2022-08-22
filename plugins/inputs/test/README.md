# Test Input Plugin

The most basic plugin that could exist.

### Configuration

```toml
[[inputs.example]]
  value = 123
```

### Metrics

- test
  - tags:
    - source
  - fields:
    - value (integer)

### Test Output

```
test,host=example.org,source=example.org value=123i 1564463260000000000
```