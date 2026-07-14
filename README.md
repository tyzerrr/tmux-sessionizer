## tmux-sessionizer
tmux-sessionizer manages your tmux-sessions.

Once you hit **tmux-sessionizer**, search directory(project) from fzf, you can start new project with new tmux session!

You can manage tmux sessions on a per-project basis!

For more details, see the demo.

And this project is inspired https://github.com/theprimeagen/tmux-sessionizer.

Thank you, ThePrimeagen.

## Usage
**tmux-sessionizer** provides five commands.
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
Displays the list of existing tmux-sessions and when you hit enter attaches the target session.

3. **tmux-sessionizer delete**
```bash
tmux-sessionizer delete
```
Displays the list of existing tmux-sessions using fzf and lets you delete them.

Selection is multi-select (`-m`): use `Tab` to mark multiple sessions, then hit enter to delete all of the selected sessions at once.

4. **tmux-sessionizer init**
```bash
tmux-sessionizer init
```
Creates the .tmux-sessionizer config file with the required `default=` prefix. If the config file already exists, it does nothing.

5. **tmux-sessionizer register**
```bash
tmux-sessionizer register <path/to/project>
```
Registers a directory as a project by appending it to the config file. The path is resolved to an absolute path before it is stored, so relative paths are safe to use.

## Demo

https://github.com/user-attachments/assets/be1d2732-38ee-41c7-9393-ecc6c0211048

## Configuration (What You Must Do)
You must write **.tmux-sessionizer**, config file.
```text
default=~/personal, ~/projects, ./ # comma separated, both absolute/relative are acceptable.
```

tmux-sessionizer searches directories and displays them using fzf, but it does not search all directories.

To specify which directories should be searched, you need to create a configuration file called .tmux-sessionizer in your home directory.

### Projects
Each path listed in the config is treated as a single project and shown in fzf as-is. Directories are **not** searched recursively — list every project directory explicitly.

To add a project, either edit the config file directly or run `tmux-sessionizer register <path/to/project>`.

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
