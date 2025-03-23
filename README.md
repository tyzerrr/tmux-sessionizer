## tmux-sessionizer
tmux-sessionizer manages your tmux-sessions.

Once you hit **tmux-sessionizer**, search directory(project) from fzf, you can start new project with new tmux session!

You can manage tmux sessions on a per-project basis!

For more details, see the demo.

And this project is inspired https://github.com/theprimeagen/tmux-sessionizer.

Thank you, ThePrimeagen.

## Usage
**tmux-sessionizer** provides two commands.
1. **tmux-sessionizer**

```bash
tmux-sessionizer
```
Launches the interactive session manager.

This command reads directories specified in the .tmux-sessionizer config file, displays them using fzf, and allows you to select one.
Once selected, tmux-sessionizer will either attach to the existing tmux session for that project or create a new one.

2. **tmux-sessionizer list**
```bash
tmux-sessionizer list
```
Displays the list of existing tmux-sessions defined and when you hit enter attach the target session.

## Demo

https://github.com/user-attachments/assets/be1d2732-38ee-41c7-9393-ecc6c0211048

## Configuration (What You Must Do)
You must write **.tmux-sessionizer**, config file.
```text
default=~/personal, ~/projects, ./ # comma separated, both absolute/relative are acceptable.
```

tmux-sessionizer searches directories and displays them using fzf, but it does not search all directories.

To specify which directories should be searched, you need to create a configuration file called .tmux-sessionizer.
This file can be created per project.

When reading configuration files, the one in the current directory takes precedence over the one in your home directory.

## Installation
You can install with homebrew.
```bash
brew install TlexCypher/tap/tmux-sessionizer
```
Also then, you can build from source.
```bash
go install github.com/TlexCypher/tmux-sessionizer
```

## Contribution
Any type of contribution is welcome.
