# workon
JIRA / Gitflow Tool to:
+ checkout a new branch from the current branch (e.g. `git workon TEAM-1234/part1` ➡️ `feature/TEAM-1234/part1`)
+ assign the corresponding JIRA Issue to yourself
+ move the corresponding JIRA Issue to `In Progress` status

By default, the branch name will be prefixed with `feature/`, but this can be overridden by 
setting the `GIT_WORKON_PREFIX` environment variable to something else. 
You can even set it to the empty string to have no prefix.

So for example, if you run:

```shell
git workon TEAM-1234
```
It will:
+ perform a `git checkout -b feature/TEAM-1234`
+ assign TEAM-1234 issue to you
+ set TEAM-1234 to "In Progress" status

You can also add a free form suffix like:

```shell
git workon TEAM-1234/part2
```

It will:
+ perform a `git checkout -b feature/TEAM-1234/part2`
+ assign TEAM-1234 issue to you (if it is not already)
+ set TEAM-1234 to "In Progress" status (if is not already in that status)




