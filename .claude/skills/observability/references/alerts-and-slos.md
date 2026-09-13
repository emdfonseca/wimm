# Alerts and SLOs

## Symptoms, not causes

An alert exists because a user is, or is about to be, affected. So alerts are defined on what users see:

- error ratio over the SLO
- latency percentile over the SLO
- availability (synthetic probe or request success rate)
- for async work: staleness or backlog age, because the user's experience is "my thing has not happened"

CPU, memory, disk, queue length, and connection counts are dashboard panels and, at most, *tickets* — not pages. A page for high CPU wakes someone to discover nothing was wrong for users.

## SLOs

Each service has one or two: availability and latency for its primary endpoints, freshness for async pipelines. Written down in the service's README with the target, window, and the owner who agreed to it.

Alert on **burn rate**, not on raw threshold: a multi-window (fast + slow) burn-rate alert catches both a sudden outage and a slow bleed, and it is quiet when the error budget is healthy. The exact rules are boilerplate; adopt the standard multi-window pattern once in `infra/` and instantiate it per SLO.

## Every alert has

```text
name         billing-availability-burn-fast
slo          billing availability 99.9% / 30d
condition    burn rate > 14.4 over 1h AND > 14.4 over 5m
severity     page | ticket
owner        the team in CODEOWNERS for apps/billing
runbook      link, with: what this means, what to check first, how to mitigate, how to escalate
```

Missing runbook → not mergeable. Missing owner → not mergeable. A page nobody owns is a page everybody ignores.

## Where they live

`infra/alerts/<service>.yaml` (or the backend's equivalent), provisioned by the same pipeline as dashboards. Alerts defined by clicking in a UI are alerts that disappear, and alerts nobody can review.

## Noise is a bug

Track alerts per week. An alert that fires and is acknowledged without action three times gets its threshold changed, gets downgraded to a ticket, or gets deleted. The goal is that a page always means "act now"; every false page erodes that until a real one is ignored.
