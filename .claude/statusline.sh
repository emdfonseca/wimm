#!/usr/bin/env bash
input=$(cat)

# ---- pastel 256-color palette (the SGR "dim" attribute muddies basic 8
# colors on dark backgrounds; soft 256-color shades stay legible instead) ----
c_reset=$'\033[0m'
c_repo=$'\033[38;5;152m'      # pastel cyan
c_branch=$'\033[38;5;151m'    # pastel green
c_ahead=$'\033[38;5;151m'     # pastel green
c_behind=$'\033[38;5;217m'    # pastel red/pink
c_dirty=$'\033[38;5;223m'     # pastel yellow
c_model=$'\033[38;5;183m'     # pastel lavender
c_effort=$'\033[38;5;153m'    # pastel blue
c_sep=$'\033[38;5;245m'       # muted gray
c_ok=$'\033[38;5;151m'        # pastel green
c_warn=$'\033[38;5;223m'      # pastel yellow
c_crit=$'\033[38;5;217m'      # pastel red/pink
c_pr_approved=$'\033[38;5;151m'
c_pr_pending=$'\033[38;5;223m'
c_pr_changes=$'\033[38;5;217m'
c_pr_draft=$'\033[38;5;245m'
c_vim=$'\033[38;5;159m'       # pastel teal
c_path=$'\033[38;5;180m'      # pastel tan
c_thinking=$'\033[38;5;210m'  # pastel salmon
c_stats=$'\033[38;5;216m'     # pastel orange
c_cost=$'\033[38;5;186m'      # pale gold

sep=" ${c_sep}│${c_reset} "

# Single jq call for every top-level field this script needs — one fork
# instead of one per field. Delimiter is \x1f (unit separator), not tab:
# bash's `read` collapses consecutive IFS-whitespace delimiters (tab counts
# as whitespace even when IFS is set to just "\t"), which would silently
# drop empty fields and shift everything after them.
IFS=$'\x1f' read -r cwd worktree_name worktree_branch git_worktree vim_mode \
  model effort thinking_enabled remaining five week five_reset week_reset \
  pr_number pr_state cost_usd <<< "$(jq -r '
  [
    (.workspace.current_dir // .cwd // ""),
    (.worktree.name // ""),
    (.worktree.branch // ""),
    (.workspace.git_worktree // ""),
    (.vim.mode // ""),
    (.model.display_name // ""),
    (.effort.level // ""),
    (.thinking.enabled // false),
    (.context_window.remaining_percentage // ""),
    (.rate_limits.five_hour.used_percentage // ""),
    (.rate_limits.seven_day.used_percentage // ""),
    (.rate_limits.five_hour.resets_at // ""),
    (.rate_limits.seven_day.resets_at // ""),
    (.pr.number // ""),
    (.pr.review_state // ""),
    (.cost.total_cost_usd // "")
  ] | map(tostring) | join("")
' <<< "$input")"

# ---- terminal width -> layout mode. Claude Code captures stdout, so there is
# no tty and `tput cols` reports terminfo's 80-column default rather than the
# real width; COLUMNS is exported by the harness (v2.1.153+) and is the only
# reliable source. Unset (older harness) fails OPEN to the widest layout —
# a too-long line wraps, but a too-terse one silently hides information.
cols=${COLUMNS:-0}
case "$cols" in ''|*[!0-9]*) cols=0 ;; esac
[ "$cols" -le 0 ] && cols=999

if   [ "$cols" -ge 110 ]; then layout=wide
elif [ "$cols" -ge 75 ];  then layout=medium
else                           layout=narrow
fi

# Resolve git branch, the *repo* name (the main repo, not the worktree dir),
# dirty file count, and ahead/behind vs upstream. Cached for 2s per git-dir so
# several sessions (or rapid refreshes of the same session) in the same repo
# don't all re-run git within the same tick.
branch=""
repo=""
repo_root=""
dirty=0
ahead=0
behind=0
if [ -n "$cwd" ] && command -v git >/dev/null 2>&1; then
  git_dir=$(git -C "$cwd" --no-optional-locks rev-parse --git-dir 2>/dev/null)
  cache_file=""
  if [ -n "$git_dir" ]; then
    case "$git_dir" in /*) ;; *) git_dir="$cwd/$git_dir" ;; esac
    cache_key=$(printf '%s' "$git_dir" | cksum | awk '{print $1}')
    cache_file="${TMPDIR:-/tmp}/claude-statusline-git-${cache_key}.cache"
  fi

  cache_hit=0
  if [ -n "$cache_file" ] && [ -f "$cache_file" ]; then
    cache_age=$(( $(date +%s) - $(date -r "$cache_file" +%s 2>/dev/null || echo 0) ))
    [ "$cache_age" -lt 2 ] && cache_hit=1
  fi

  if [ "$cache_hit" -eq 1 ]; then
    IFS='|' read -r branch repo repo_root dirty ahead behind < "$cache_file"
  else
    branch=$(git -C "$cwd" --no-optional-locks branch --show-current 2>/dev/null)
    # git-common-dir points at the shared .git (the main worktree's), so its
    # parent is the real repo root even when cwd is a linked worktree.
    gcd=$(git -C "$cwd" --no-optional-locks rev-parse --git-common-dir 2>/dev/null)
    if [ -n "$gcd" ]; then
      case "$gcd" in /*) ;; *) gcd="$cwd/$gcd" ;; esac
      # git-common-dir can come back with unresolved ".." segments (e.g. run
      # from a subdirectory) — resolve to a physical path before taking basename.
      repo_root=$(cd "$(dirname "$gcd")" 2>/dev/null && pwd)
      repo=$(basename "$repo_root")
    fi
    dirty=$(git -C "$cwd" --no-optional-locks status --porcelain 2>/dev/null | wc -l | tr -d ' ')
    # rev-list walks commit history; skip it entirely when there's no upstream to diff against.
    if git -C "$cwd" --no-optional-locks rev-parse --verify --quiet '@{upstream}' >/dev/null 2>&1; then
      ab=$(git -C "$cwd" --no-optional-locks rev-list --left-right --count '@{upstream}...HEAD' 2>/dev/null)
      if [ -n "$ab" ]; then
        behind=$(echo "$ab" | awk '{print $1}')
        ahead=$(echo "$ab" | awk '{print $2}')
      fi
    fi
    [ -n "$cache_file" ] && printf '%s|%s|%s|%s|%s|%s' "$branch" "$repo" "$repo_root" "$dirty" "$ahead" "$behind" > "$cache_file" 2>/dev/null
  fi
fi

# Prefer worktree branch if available
[ -z "$branch" ] && [ -n "$worktree_branch" ] && branch="$worktree_branch"

# Fall back to the cwd basename if git couldn't tell us the repo name
[ -z "$repo" ] && repo="$(basename "$cwd")"

# ---- workspace-relative path (cwd relative to repo root, last 1-2 path
# segments) — e.g. "apps/extension" when cwd is the extension package root,
# so a pnpm-workspace monorepo shows which package you're in at a glance. ----
path_seg=""
if [ -n "$repo_root" ] && [ -n "$cwd" ] && [ "$cwd" != "$repo_root" ]; then
  case "$cwd" in
    "$repo_root"/*)
      rel_path="${cwd#"$repo_root"/}"
      IFS='/' read -ra rel_segs <<< "$rel_path"
      n=${#rel_segs[@]}
      if [ "$n" -ge 2 ]; then
        path_seg="${rel_segs[$((n-2))]}/${rel_segs[$((n-1))]}"
      elif [ "$n" -eq 1 ]; then
        path_seg="${rel_segs[0]}"
      fi
      ;;
  esac
fi

parts=()

# Repo
if [ -n "$repo" ]; then
  parts+=("${c_repo}${repo}${c_reset}")
fi

# Branch / worktree — show both only when they actually differ — plus
# dirty-file count and ahead/behind-upstream indicators.
branch_seg=""
# Icons cost 2 display columns each and carry no information the name doesn't —
# first thing to go when narrow.
if [ "$layout" = narrow ]; then b_icon=""; w_icon=""; else b_icon="🌿 "; w_icon="🪵 "; fi

# A long branch name ("feat/some-very-descriptive-slug") can blow any width
# budget on its own — cap it when narrow and mark the elision.
trunc_branch() {
  local s="$1" max=18
  if [ "$layout" = narrow ] && [ "${#s}" -gt "$max" ]; then
    printf '%s…' "${s:0:$((max - 1))}"
  else
    printf '%s' "$s"
  fi
}
branch=$(trunc_branch "$branch")
if [ -n "$worktree_name" ] || [ -n "$git_worktree" ]; then
  label="${worktree_name:-$git_worktree}"
  # Treat a "worktree-"-prefixed branch as the same as its worktree name.
  if [ -n "$branch" ] && [ "$branch" != "$label" ] && [ "$branch" != "worktree-$label" ]; then
    branch_seg="${b_icon}${branch} (${w_icon}${label})"
  else
    # branch == worktree (or just the prefixed form): one worktree icon, clean name
    branch_seg="${w_icon}${label}"
  fi
elif [ -n "$branch" ]; then
  branch_seg="${b_icon}${branch}"
fi

if [ -n "$branch_seg" ]; then
  # Pack ahead/behind/dirty with no inner spaces: "↑167*1", not " ↑167 *1".
  state=""
  [ "$ahead" -gt 0 ] && state="${state}${c_ahead}↑${ahead}${c_reset}"
  [ "$behind" -gt 0 ] && state="${state}${c_behind}↓${behind}${c_reset}"
  [ "$dirty" -gt 0 ] && state="${state}${c_dirty}*${dirty}${c_reset}"
  [ -n "$state" ] && state=" ${state}"
  parts+=("${c_branch}${branch_seg}${c_reset}${state}")
fi

# Workspace-relative path (wide/medium only)
[ -n "$path_seg" ] && [ "$layout" != narrow ] && parts+=("${c_path}${path_seg}${c_reset}")

# ---- vim mode (only present when vim keybindings are enabled) ----
[ -n "$vim_mode" ] && parts+=("${c_vim}[${vim_mode}]${c_reset}")

# ---- model + effort level + thinking indicator ----
# display_name carries a parenthetical variant suffix ("Opus 4.8 (1M context)")
# that eats ~13 columns and never changes within a session — strip it, then
# abbreviate the family to its initial: "Opus 4.8" -> "O4.8", "Sonnet 5" -> "S5".
if [ -n "$model" ]; then
  model_short=$(sed -E 's/ *\([^)]*\)//g; s/^Claude +//; s/^Opus +/O/; s/^Sonnet +/S/; s/^Haiku +/H/' <<< "$model")
  seg="${c_model}${model_short}${c_reset}"
  if [ -n "$effort" ]; then
    case "$effort" in
      low) e="L" ;; medium) e="M" ;; high) e="H" ;;
      xhigh) e="X" ;; max) e="!" ;; *) e="${effort:0:1}" ;;
    esac
    seg="${seg}${c_sep}·${c_reset}${c_effort}${e}${c_reset}"
  fi
  [ "$thinking_enabled" = "true" ] && [ "$layout" != narrow ] && seg="${seg} ${c_thinking}∴${c_reset}"
  parts+=("$seg")
fi

# ---- context window remaining % (green ok / yellow low / red critical) ----
if [ -n "$remaining" ]; then
  remaining_r=$(printf '%.0f' "$remaining")
  if [ "$remaining_r" -ge 50 ]; then c="$c_ok"
  elif [ "$remaining_r" -ge 20 ]; then c="$c_warn"
  else c="$c_crit"
  fi
  # Narrow: the colour already encodes ok/low/critical, so the "ctx:" label is
  # redundant — drop it and keep the number.
  if [ "$layout" = narrow ]; then
    parts+=("${c}${remaining_r}%${c_reset}")
  else
    parts+=("${c}ctx:${remaining_r}%${c_reset}")
  fi
fi

# ---- session cost (metered-equivalent; on a subscription this is the
# counterfactual API spend, not a bill — the rate-limit segment below is the
# budget that actually binds, which is why cost is the first thing dropped) ----
if [ -n "$cost_usd" ] && [ "$layout" = wide ]; then
  parts+=("${c_cost}\$$(printf '%.2f' "$cost_usd")${c_reset}")
fi

# ---- rate limits (Claude.ai subscription: 5h / 7d used %) ----
usage_color() {
  awk -v p="$1" 'BEGIN { if (p >= 80) print "crit"; else if (p >= 50) print "warn"; else print "ok" }'
}

# Projects whether the current consumption rate will exhaust a rate-limit
# window before it resets. The API reports used_percentage at ~1-point
# resolution, so a rate sampled across two nearby renders is quantisation
# noise, not a trend: one 3%->4% tick 4 minutes apart reads as 0.25%/min and
# "exhausted in 6h" on a window with days left. Two guards instead:
#   * the trailing rate is measured against the newest sample at least
#     min_span old (window/20, floor 15m), and only trusted once usage has
#     moved >= 2 points over that span, keeping quantisation error under 50%;
#   * samples are keyed to resets_at, so a window rollover discards the
#     history rather than folding a negative delta into the estimate.
# The reported figure is still max(since-window-start average, trailing rate)
# so a steady burn and a fresh acceleration both surface. Prints
# "<eta> <projected_pct_at_reset>"; eta is empty unless the window is
# projected to hit 100% before it resets.
rate_eta() {
  local pct="$1" resets_at="$2" cache_file="$3" window_seconds="$4"
  [ -z "$pct" ] || [ -z "$resets_at" ] && return
  awk -v pct="$pct" -v resets_at="$resets_at" -v now="$(date +%s)" -v cache="$cache_file" -v window_seconds="$window_seconds" '
    BEGIN {
      pct += 0; resets_at += 0; window_seconds += 0

      # History is "<resets_at> <t>:<pct> <t>:<pct> ...". A resets_at that
      # does not match the live one belongs to a window that has since
      # rolled over; its samples describe a counter that no longer exists.
      n = 0
      if ((getline line < cache) > 0) {
        m = split(line, f, " ")
        if (m >= 1 && (f[1] + 0) == resets_at) {
          for (i = 2; i <= m; i++) {
            if (split(f[i], kv, ":") == 2) { n++; st[n] = kv[1] + 0; sp[n] = kv[2] + 0 }
          }
        }
      }
      close(cache)

      # Append-only, at most one sample per minute. Never refresh the newest
      # in place: a creeping timestamp means the history never ages and no
      # sample is ever old enough to anchor a trailing rate.
      if (n == 0 || now - st[n] >= 60) { n++; st[n] = now; sp[n] = pct }

      first = 1
      while (first <= n && now - st[first] > window_seconds) first++
      if (n - first + 1 > 240) first = n - 239

      out = resets_at
      for (i = first; i <= n; i++) out = out " " st[i] ":" sp[i]
      tmp = cache ".tmp"
      print out > tmp
      close(tmp)
      system("mv -f " tmp " " cache)

      min_span = window_seconds / 20
      if (min_span < 900) min_span = 900

      trailing_rate = 0
      for (i = n; i >= first; i--) {
        if (now - st[i] >= min_span) {
          span = now - st[i]
          delta = pct - sp[i]
          if (delta >= 2) trailing_rate = delta / span
          break
        }
      }

      avg_rate = 0
      elapsed_in_window = window_seconds - (resets_at - now)
      if (elapsed_in_window > 0) avg_rate = pct / elapsed_in_window

      effective_rate = (avg_rate > trailing_rate) ? avg_rate : trailing_rate

      remaining = resets_at - now
      if (remaining < 0) remaining = 0
      projected = pct + effective_rate * remaining
      if (projected > 999) projected = 999

      eta = "-"
      if (effective_rate > 0 && projected >= 100) {
        seconds_to_100 = (100 - pct) / effective_rate
        if (seconds_to_100 < 3600) eta = sprintf("%.0fm", seconds_to_100 / 60)
        else eta = sprintf("%.1fh", seconds_to_100 / 3600)
      }
      printf "%s %.0f", eta, projected
    }
  '
}

# Narrow drops the 5h:/7d: labels and the % signs ("31/58"); the exhaustion ETA
# is kept in every layout — it is the one segment that is actionable, and it
# only renders when a window is actually projected to run out.
if [ "$layout" = narrow ]; then l5=""; l7=""; pc=""; rl_join="${c_sep}/${c_reset}"
else                            l5="5h:"; l7="7d:"; pc="%"; rl_join=" "
fi

# Time left in a window, in the fewest characters that stay unambiguous:
# minutes under an hour ("47m"), one decimal hour above it ("4.9h").
time_left() {
  [ -z "$1" ] && return
  awk -v resets_at="$1" -v now="$(date +%s)" 'BEGIN {
    left = resets_at - now
    if (left < 0) left = 0
    if (left < 3600) printf "%.0fm", left / 60
    else printf "%.1fh", left / 3600
  }'
}

# Renders what rate_eta found: the projected usage at reset once it is close
# enough to matter (wide only — it costs columns), and the exhaustion ETA,
# which is the actionable half and so survives every layout.
rate_outlook() {
  local eta="$1" proj="$2" out=""
  [ "$layout" = wide ] && [ -n "$proj" ] && [ "$proj" -ge 70 ] 2>/dev/null \
    && out="${c_sep}→${c_reset}$([ "$proj" -ge 100 ] && printf '%s' "$c_crit" || printf '%s' "$c_warn")${proj}%${c_reset}"
  [ -n "$eta" ] && [ "$eta" != "-" ] && out="${out} ${c_crit}⚠~${eta}${c_reset}"
  printf '%s' "$out"
}

rate_seg=""
if [ -n "$five" ]; then
  case "$(usage_color "$five")" in ok) c="$c_ok";; warn) c="$c_warn";; crit) c="$c_crit";; esac
  rate_seg="${c}${l5}$(printf '%.0f' "$five")${pc}${c_reset}"
  if [ -n "$five_reset" ]; then
    read -r eta proj <<< "$(rate_eta "$five" "$five_reset" "${TMPDIR:-/tmp}/claude-statusline-ratelimit-five.cache" 18000)"
    rate_seg="${rate_seg}$(rate_outlook "$eta" "$proj")"
    # Time to the session reset: the 5h window is the one that bites within a
    # sitting, and "wait it out" is only a choice if you know how long.
    if [ "$layout" != narrow ]; then
      left=$(time_left "$five_reset")
      [ -n "$left" ] && rate_seg="${rate_seg}${c_sep}·${left}${c_reset}"
    fi
  fi
fi
if [ -n "$week" ]; then
  case "$(usage_color "$week")" in ok) c="$c_ok";; warn) c="$c_warn";; crit) c="$c_crit";; esac
  [ -n "$rate_seg" ] && rate_seg="${rate_seg}${rl_join}"
  rate_seg="${rate_seg}${c}${l7}$(printf '%.0f' "$week")${pc}${c_reset}"
  if [ -n "$week_reset" ]; then
    read -r eta proj <<< "$(rate_eta "$week" "$week_reset" "${TMPDIR:-/tmp}/claude-statusline-ratelimit-week.cache" 604800)"
    rate_seg="${rate_seg}$(rate_outlook "$eta" "$proj")"
  fi
fi
[ -n "$rate_seg" ] && parts+=("$rate_seg")

# ---- open PR for current branch (colour carries the review state when narrow) ----
if [ -n "$pr_number" ] && [ "$layout" != narrow ]; then
  case "$pr_state" in
    approved) c="$c_pr_approved" ;;
    changes_requested) c="$c_pr_changes" ;;
    draft) c="$c_pr_draft" ;;
    *) c="$c_pr_pending" ;;
  esac
  label="PR #${pr_number}"
  [ -n "$pr_state" ] && [ "$layout" = wide ] && label="${label} (${pr_state})"
  parts+=("${c}${label}${c_reset}")
fi

# Join with a dim separator
out=""
for p in "${parts[@]}"; do
  if [ -z "$out" ]; then
    out="$p"
  else
    out="${out}${sep}${p}"
  fi
done

printf '%s' "$out"
