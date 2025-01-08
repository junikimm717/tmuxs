# tmuxs

A tmux sessionizer rewritten in go, so the program works more consistently
across the different operating systems I use.

Configuration is quite limited because this program is mostly meant for personal
use. The fuzzy finder does not provide directories in deterministic order
because they are being searched in parallel as the fuzzy finder is loaded.

## Setup

1. Install the binary. `go install github.com/junikimm717/tmuxs` works.
2. Set the `WORKSPACES` environment variable as a colon-separated list of paths
   that you want to be searched by the fuzzy finder, e.g.
   `/home/junikim/programs:/home/junikim/schoolwork`.
