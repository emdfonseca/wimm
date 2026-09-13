# CI

CI runs `devbox run -- just <verb>` and nothing else. Every step corresponds to a recipe a developer can run locally with the same result.

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: jetify-com/devbox-install-action@v0.12.0
      - run: devbox run -- just ci
```
