---
name: git-commit
description: Review local git changes, write an appropriate commit message, create commits, and push the current branch to every remote. Use when Codex needs to handle repository submission work such as checking `git status`, summarizing diffs, committing staged or unstaged changes, committing and pushing git submodules before the parent repository, then committing the parent repository with the updated submodule pointer and pushing the active branch to all remotes.
---

# Git Commit

## Overview

Use this skill for repository submission work. Inspect changes first, derive a commit message from the actual diff, commit and push submodules first when they are involved, then commit the parent repository with the updated submodule pointer and push the current branch to every remote.

## Workflow

1. Read repository state before changing anything.
2. Detect submodules and handle them before touching the parent repository commit.
3. For each affected submodule, commit and push the submodule repository first.
4. Return to the parent repository and commit the parent repository with the updated submodule pointer plus any related parent-repository changes.
5. Push the current parent-repository branch to every remote after the parent commit succeeds.

## Inspect Repository State

Run these checks up front:

```bash
git status --short
git branch --show-current
git remote -v
git submodule status
```

Inspect the actual diff before writing a commit message:

```bash
git diff --stat
git diff --submodule=log
```

If the worktree is clean, report that there is nothing to commit. Do not create an empty commit unless the user explicitly asks for one.

## Handle Submodules First

Treat submodules as independent repositories, but always finish them before the parent repository.

- If `git status --short` in the parent repo shows a changed submodule entry, inspect whether the submodule itself has uncommitted work.
- If the parent repository shows a changed submodule pointer, enter the submodule even when its worktree is clean.
- For each dirty submodule, enter the submodule and run the same inspection flow there first.
- Write a commit message for the submodule based on its own diff, not the parent repository diff.
- Commit and push the submodule before committing the parent repository.
- If the submodule worktree is clean but the parent pointer changed, verify that the submodule `HEAD` commit exists on the submodule remotes and push it if needed before committing the parent repository.
- Return to the parent repository and verify that the parent diff now contains the updated submodule pointer, and include that pointer update in the parent commit.

Use commands like:

```bash
git -C <submodule-path> status --short
git -C <submodule-path> diff --stat
git -C <submodule-path> branch --show-current
git -C <submodule-path> remote -v
git -C <submodule-path> rev-parse HEAD
git -C <submodule-path> ls-remote --heads <remote>
git -C <submodule-path> add -A
git -C <submodule-path> commit -m "<message>"
git -C <submodule-path> push <remote> <branch>
```

If a submodule has multiple remotes, push its current branch to every remote exactly like the parent repository.

Before the parent repository commit, explicitly validate the submodule remote state:

- Read the submodule `HEAD` commit with `git -C <submodule-path> rev-parse HEAD`.
- Read the submodule current branch with `git -C <submodule-path> branch --show-current`.
- Compare against each submodule remote branch with `git -C <submodule-path> rev-list <remote>/<branch>..HEAD`.
- If the branch has outgoing commits, push that submodule branch to every remote before touching the parent repository commit.
- If the submodule is in detached `HEAD`, stop and report it unless the user explicitly asks to push a detached commit reference.

After all affected submodules are pushed, re-check the parent repository diff and make sure the parent commit includes:

- The submodule pointer update for each affected submodule.
- Any related parent-repository files that should travel with that submodule change.
- No unrelated user changes unless the user explicitly asks for a broader commit.

## Write Commit Messages

Base the message on the diff, not on filenames alone.

- Prefer a short imperative subject.
- Keep the subject scoped to the dominant change.
- Add a body when the diff contains multiple related updates that need context.
- If the user asks for a comment or annotation, treat that as the commit message request and still ground it in the diff.

Good patterns:

- `docs: fix design spec links`
- `feat(ticket): add workbench redesign docs`
- `refactor(api): simplify auth middleware wiring`

## Commit Safely

Before committing, confirm the change set is understood.

- Use `git add` only for the files intended for this commit.
- Do not revert unrelated user changes.
- Do not amend an existing commit unless the user explicitly asks.
- If there are unrelated dirty files and the target commit should stay focused, stage only the intended files instead of using a broad `git add -A`.
- When submodules are involved, the parent repository commit must include the updated submodule pointer after the submodule push succeeds.

Typical commands:

```bash
git add <paths...>
git commit -m "<subject>" -m "<optional body>"
```

## Push Every Remote

After local commits succeed, push the current branch to every push remote.

- Get the active branch with `git branch --show-current`.
- Enumerate remotes with `git remote -v`.
- Deduplicate remote names.
- Push the same branch to every remote.
- Prefer parallel pushes when remotes are independent, so one slow remote does not block the others.

Typical pattern:

```bash
git push <remote> <current-branch>
```

If one remote succeeds and another fails, report the partial result clearly and include the failing remote plus the error.

## Failure Handling

- If authentication or network access fails on a remote, keep the successful pushes and report which remotes still need retry.
- If a submodule push fails, or if the referenced submodule commit is not confirmed on the submodule remotes, do not commit the parent repository submodule pointer unless the user explicitly asks to proceed with that inconsistent state.
- If the parent repository has no changes after submodule processing, report that explicitly.
- If a command hangs during remote access, retry with a non-interactive or bounded-timeout variant to surface a concrete error.
- Do not push the parent repository before the parent commit that contains the updated submodule pointer is created.

## Completion Checklist

- Submodule worktrees inspected
- Dirty submodules committed and pushed first
- Changed submodule pointers validated against submodule remotes
- Parent repository diff re-checked after submodule updates
- Parent repository commit includes the updated submodule pointer
- Parent repository committed with a diff-based message
- Current branch pushed to every remote
- Final status and any failed remotes reported back to the user
