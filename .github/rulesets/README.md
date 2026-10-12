# Repository rulesets

Portable definitions for [GitHub repository rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets) on `mlaccetti/harness-bridge`.

| File | Purpose |
|------|---------|
| `protect-main.json` | `main`: PR + CODEOWNERS review, CI job **`check`**, no force-push or delete |
| `admin-only-branch-creation.json` | All refs: only admins may create branches on origin (admin role bypass) |

These match the active rulesets in the repository (IDs are assigned by GitHub when created).

## Apply or recreate

From a clone with `gh` authenticated as a repo admin:

```bash
repo=mlaccetti/harness-bridge

gh api "repos/${repo}/rulesets" --method POST --input .github/rulesets/protect-main.json
gh api "repos/${repo}/rulesets" --method POST --input .github/rulesets/admin-only-branch-creation.json
```

To update an existing ruleset, list IDs then `PUT` with the same JSON shape:

```bash
gh api "repos/${repo}/rulesets" --jq '.[] | {id, name}'
gh api "repos/${repo}/rulesets/<id>" --method PUT --input .github/rulesets/protect-main.json
```

**Note:** `actor_id` `5` with `actor_type` `RepositoryRole` is the repository **Administrator** role (bypass for branch creation).
