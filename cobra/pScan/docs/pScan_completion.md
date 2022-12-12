## pScan completion

Generate bash or zsh completion for pScan

### Synopsis

To load your pScan completion run
source <(pScan completion bash)
source <(pScan completion zsh)

To load pScan completion automatically on login, add this line to your .bashrc file:
$ ~/.bashrc
source <(pScan completion bash)
$ ~/.zshrc
source <(pScan completion zsh)


```
pScan completion <bash|zsh> [flags]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.pScan.yaml)
  -f, --hosts-file string   pScan hosts file (default "pScan.hosts")
```

### SEE ALSO

* [pScan](pScan.md)	 - Fast TCP port scanner

###### Auto generated by spf13/cobra on 12-Dec-2022