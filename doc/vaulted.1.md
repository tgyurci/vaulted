vaulted 1
=========

NAME
----

vaulted - spawn sessions from securely stored secrets

SYNOPSIS
--------

`vaulted` `-n` *name* [`-i`]  
`vaulted` `-n` *name* [`--`] *CMD*

`vaulted` *COMMAND* [*args...*]

DESCRIPTION
-----------

If no *COMMAND* is provided, `vaulted` either spawns *CMD* (if provided) or
spawns an interactive shell.

`--` may be used to differentiate the *CMD* from `vaulted`'s own arguments.

COMMANDS
--------

`add` / `create` / `new`
  Interactively creates the content of a new vault. See vaulted-add(1).

`cp` / `copy`
  Copies the content of a vault and saves it as a new vault with a new password. See vaulted-cp(1).

`dump`
  Writes the content of a vault to stdout as JSON. See vaulted-dump(1).

`edit`
  Interactively edits the content of an existing vault. See vaulted-edit(1).

`env`
  Outputs shell commands that load secrets for a vault into the shell. See vaulted-env(1).

`exec`
  Executes shell commands with a given vault or role. See vaulted-exec(1).

`load`
  Uses JSON provided to stdin to create or replace the content of a vault. See vaulted-load(1).

`ls` / `list`
  Lists all vaults. See vaulted-ls(1).

`passwd` / `password`
  Changes the password for an existing vault. See vaulted-passwd(1).

`rm` / `delete` / `remove`
  Removes existing vaults. See vaulted-rm(1).

`shell`
  Starts an interactive shell with the secrets for the vault loaded into the shell. See vaulted-shell(1).

`upgrade`
  Upgrades legacy vaults to the current vault format. See vaulted-upgrade(1).

FILE LOCATIONS
--------------

Vaults and cached sessions are stored according to the [XDG Base Directory Specification][xdg].

**Vault** files are stored in:

* `$XDG_DATA_HOME/vaulted/` _(typically `~/.local/share/vaulted/`)_
* `$XDG_DATA_DIRS/vaulted/` _(typically `/usr/local/share` and `/usr/share`)_

Vault files are written to `$XDG_DATA_HOME/vaulted/`. To backup your Vaulted data, all files in
this directory should be backed up. Session cache files do not need to be retained.

**Session** cache files are stored in:

* `$XDG_CACHE_HOME/vaulted/` _(typically `~/.cache/vaulted/`)_

[xdg]: https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html

EXIT CODES
----------

|Exit code|Meaning|
|:-:|---|
| 0 | Success. |
| 64 | Invalid CLI usage (see message for more details). |
| 65 | There was an unrecoverable problem with the vault file. |
| 69 | A required service is presently unavailable (e.g. askpass). |
| 79 | Invalid password supplied. |

GUI Password Prompts
--------------------

Although Vaulted tries to make sure you can redirect `stdin` and friends,
sometimes it is still preferable to use a GUI-based password prompt. For this
reason, Vaulted can be configured to use an askpass implementation. Vaulted's
askpass integration is triggered when the `VAULTED_ASKPASS` variable is set.

Pointing `VAULTED_ASKPASS` to an executable file that implements askpass will
cause Vaulted to use execute the file specified to prompt the user for
passwords. The first parameter provided to the executable is prompt text
intended to be shown to the user. The askpass implementation then writes the
password to `stdout` and returns a success code (0). If a failure code (non-0)
is returned, the password input is aborted.

The vault name, requested secret type (password, MFA token etc.) and password
request reason is passed to the askpass process in the environment variables
`VAULTED_ENV`, `VAULTED_PASSWORD_TYPE` and `VAULTED_PASSWORD_REASON`
respectively.

Valid values for `VAULTED_PASSWORD_TYPE` are: `password`, `legacypassword` or
`mfatoken`.

Valid values for `VAULTED_PASSWORD_REASON` are: `new`, `nomatch`, `confirm` or
the empty string if `VAULTED_PASSWORD_TYPE` is not `password`.

Vaulted is intended to integrate seamlessly with existing askpass
implementations (e.g. `ssh-askpass`).

On macOS, a simple AppleScript askpass implementation can be used:

```AppleScript
#!/usr/bin/env osascript

on run argv
    if length of argv > 0 then
        set message to item 1 of argv
    else
        set message to "Password:"
    end if

    set frontmost_application to name of (info for (path to frontmost application))
    tell application frontmost_application
        display dialog message with title "Vaulted" with icon caution with hidden answer default answer ""

        text returned of result
    end tell
end run
```
