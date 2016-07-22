package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/miquella/vaulted/lib"
	"github.com/miquella/vaulted/lib/legacy"
	"github.com/spf13/pflag"
)

func main() {
	// omit the command name that is passed to VaultedCLI
	var cli VaultedCLI
	if len(os.Args) > 0 {
		cli = VaultedCLI(os.Args[1:])
	}

	cli.Run()
}

func getPassword() string {
	password, err := speakeasy.Ask("Password: ")
	if err != nil {
		os.Exit(1)
	}
	return password
}

func openVault(name string) (password string, vault *vaulted.Vault, err error) {
	password = os.Getenv("VAULTED_PASSWORD")
	if password != "" {
		vault, err = vaulted.OpenVault(password, name)
	} else {
		for i := 0; i < 3; i++ {
			password = getPassword()
			vault, err = vaulted.OpenVault(password, name)
			if err != vaulted.ErrInvalidPassword {
				break
			}
		}
	}
	return
}

func openLegacyVault() (password string, environments map[string]legacy.Environment, err error) {
	legacyVault, err := legacy.ReadVault()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	password = os.Getenv("VAULTED_PASSWORD")
	if password != "" {
		environments, err = legacyVault.DecryptEnvironments(password)
	} else {
		for i := 0; i < 3; i++ {
			password = getPassword()
			environments, err = legacyVault.DecryptEnvironments(password)
			if err != legacy.ErrInvalidPassword {
				break
			}
		}
	}
	return
}

type VaultedCLI []string

func (cli VaultedCLI) Run() {
	if len(cli) == 0 {
		os.Exit(255)
	}

	switch cli[0] {
	case "cat":
		cli.Cat()

	case "list", "ls":
		cli.List()

	case "rm":
		cli.Remove()

	case "shell":
		cli.Shell()

	case "upgrade":
		cli.Upgrade()

	default:
		if strings.HasPrefix(cli[0], "-") {
			cli.Spawn()
		} else {
			os.Exit(255)
		}
	}
}

func (cli VaultedCLI) Cat() {
	if len(cli) != 2 {
		fmt.Fprintln(os.Stderr, "You must specify a single vault to cat")
		os.Exit(255)
	}

	_, vault, err := openVault(cli[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	staticVars, err := vault.CreateEnvironment(true, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var keys []string
	for key, _ := range staticVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("%s=%s", key, staticVars[key]))
	}
}

func (cli VaultedCLI) List() {
	vaults, err := vaulted.ListVaults()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Failed to list vaults: %v", err))
		os.Exit(1)
	}

	for _, vault := range vaults {
		fmt.Fprintln(os.Stdout, vault)
	}
}

func (cli VaultedCLI) Remove() {
	if len(cli) <= 1 {
		fmt.Fprintln(os.Stderr, "You must specify which vaults to remove")
		os.Exit(255)
	}

	failures := 0
	for _, name := range cli[1:] {
		err := vaulted.RemoveVault(name)
		if err != nil {
			failures++
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %v", name, err))
		}
	}

	os.Exit(failures)
}

func (cli VaultedCLI) Shell() {
	if len(cli) != 2 {
		fmt.Fprintln(os.Stderr, "You must specify a single vault to spawn a shell with")
		os.Exit(255)
	}

	_, vault, err := openVault(cli[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code, err := vault.Spawn([]string{os.Getenv("SHELL"), "--login"}, map[string]string{"VAULTED_ENV": cli[1]})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(*code)
}

func (cli VaultedCLI) Spawn() {
	spawnFlags := pflag.NewFlagSet("spawn", pflag.ContinueOnError)
	spawnFlags.SetInterspersed(false)

	name := spawnFlags.StringP("name", "n", "", "Name of the vault to spawn")
	interactive := spawnFlags.BoolP("interactive", "i", false, "Spawn an interactive shell")
	err := spawnFlags.Parse([]string(cli))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(255)
	}

	if spawnFlags.ArgsLenAtDash() > 0 {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Unknown argument(s): %v", spawnFlags.Args()[:spawnFlags.ArgsLenAtDash()]))
		os.Exit(255)
	}

	_, vault, err := openVault(cli[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var cmd []string
	if *interactive {
		cmd = append(cmd, os.Getenv("SHELL"), "--login")
	}
	cmd = append(cmd, spawnFlags.Args()...)

	code, err := vault.Spawn(cmd, map[string]string{"VAULTED_ENV": *name})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(*code)
}

func (cli VaultedCLI) Upgrade() {
	password, environments, err := openLegacyVault()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// collect the current list of vaults (so we don't overwrite any)
	vaults, _ := vaulted.ListVaults()
	existingVaults := map[string]bool{}
	for _, name := range vaults {
		existingVaults[name] = true
	}

	failed := 0
	for name, env := range environments {
		if existingVaults[name] {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: skipped (vault already exists)", name))
			continue
		}

		vault := vaulted.Vault{
			Vars: env.Vars,
		}
		err = vaulted.SealVault(password, name, &vault)
		if err != nil {
			failed++
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %v", name, err))
		} else {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: upgraded", name))
		}
	}

	os.Exit(failed)
}