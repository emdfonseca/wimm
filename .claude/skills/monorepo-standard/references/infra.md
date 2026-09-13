# Infrastructure

## Layout

```text
infra/
├── modules/            # reusable Terraform modules, versioned with the repo
│   ├── service/
│   └── database/
├── envs/
│   ├── dev/
│   ├── staging/
│   └── prod/
└── justfile
```

Environments are directories, not workspaces or variable files. A directory per environment makes the diff between prod and staging a readable file comparison, and it makes it impossible to apply the wrong environment by forgetting a flag.

## What infra may reference

Built artifacts — image tags, binary versions, bundle hashes — never application source. Infra that reaches into `apps/*/src` couples deployment to code layout, and the coupling only surfaces when a refactor breaks a deploy.

Application config that changes per environment belongs to infra. Config that is the same everywhere belongs with the app.

## Tasks

```just
plan env:
    terraform -chdir=envs/{{env}} plan

apply env:
    terraform -chdir=envs/{{env}} apply

fmt:
    terraform fmt -recursive

lint:
    terraform fmt -check -recursive
    tflint --recursive

check: lint
    just plan dev
```

Keep `apply` out of any aggregate recipe. `just check` at the root must stay safe to run on any machine at any time — the moment it can change infrastructure, people stop running it.

## State and secrets

Remote state with locking, configured per environment. No credentials in the repo, including in `devbox.json`; devbox declares the CLI, the operator supplies the credentials. Document which identity is expected in `infra/README.md` rather than in someone's shell history.
