# Omlox Hub CLI

Omlox Hub CLI client.
You can check the CLI reference documentation [here](./cli/omlox.md).

> [!NOTE]  
> The currently being developed.
> Please try it out and reach out to us with feedback and features you would like see included. 

1. [Install](#install)
   - [From source](#from-source)
   - [Add completions](#add-completions)
     - [Example for `zsh`](#example-for-zsh)
     - [Setup your shell to run go installed binaries](#setup-your-shell-to-run-go-installed-binaries)

## Install

### From source

First, you will require a few tools for this process:

- `go>=1.21`
- `make`
- `git`

Then you can procede with the installation: 

1. Clone the project to your computer
2. Run `make install`

That should be it. Now try to run the CLI:

```console
omlox version
```

If it says something like "command not found", it may be because you do not have your `PATH` setup for go binaries.
Your can run through the setup at [Setup your shell to run go installed binaries](#setup-your-shell-to-run-go-installed-binaries).

### Add completions

There are multiple ways of adding auto-completions.
This is the easier of all (may not the best).

To activate the autocompletion in your current shell:

```console
source <(omlox completion your-shell) # zsh, bash, etc.
```

#### Example for `zsh`

```console
source <(omlox completion zsh)
```

To permanently load the auto completions on new shell, you must the previous command in your shell config file. Add it like soo:

```console
echo "source <(omlox completion zsh)" >> ~./zshrc
```

### Setup your shell to run go installed binaries

Let's go over how you can setup your shell to run go programs that you have installed.

First, check if you have your `GOPATH`:

```console
echo $GOPATH
```

If nothing shows up, you must define it.
To do soo, you will need to add some things to your shell configuration file.

To check which shell your are using, run:

```console
echo $0
```

For zsh you will have to change the `~/.zshrc` file, and for bash you will have to change `~/.bashrc`, etc..

Here I will be using zsh as example, be sure to change the config file to suit your shell.

Set the `GOPATH` to the directory of your go installation.
In my computer it's at `~/go`, so:

```console
echo "export GOPATH=\$HOME/go" >> ~/.zshrc
```

Now add the go binaries to your `PATH`:

```console
echo "export PATH=\$GOPATH/bin:\$PATH" >> ~/.zshrc
```

Open a new shell and you should be able to run your go installed binaries.
