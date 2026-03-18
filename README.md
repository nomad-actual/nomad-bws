# nomad-bws
Bitwarden secrets manager provider for Nomad

todo: improve :)

# Setup
Install go, gcc, build-essentials

Pull Repo
cd repo

```
go build -o nomad-bws main.go
```

cp nomad-bws to your Nomad plugins secrets dir
Usually this is
`/opt/nomad/common_plugins/secrets`

```
sudo install -m 0755 nomad-bws /opt/nomad/common_plugins/secrets/nomad-bws
```

Verify
```
/opt/nomad/common_plugins/secrets/nomad-bws fingerprint
```

Expected:
```
{"type":"secrets","version":"0.1.0"}
```
