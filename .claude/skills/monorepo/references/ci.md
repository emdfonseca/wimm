# CI

CI runs `devbox run -- just <verb>` and nothing else. Every step corresponds to a recipe a developer can run locally with the same result.

The workflow (`.github/workflows/ci.yml`) is three steps: checkout, install devbox, run that one command. Tool versions come from `devbox.json`, never from the YAML.
